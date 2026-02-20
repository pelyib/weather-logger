package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/pelyib/weather-logger/adapter/provider/accuweather"
	"github.com/pelyib/weather-logger/adapter/provider/openweather"
	"github.com/pelyib/weather-logger/adapter/sqlite"
	"github.com/pelyib/weather-logger/adapter/web"
	"github.com/pelyib/weather-logger/domain"
	"github.com/pelyib/weather-logger/internal/config"
	portout "github.com/pelyib/weather-logger/port/out"
)

func main() {
	runOnce := flag.String("run-once", "", "fetch data and exit: 'forecasts' or 'historical'")
	flag.Parse()

	logger := slog.Default()

	cnf, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	repo, err := sqlite.Open(cnf.Database.Path)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer repo.Close()

	providers := buildProviders(cnf, logger)
	locations := toDomainLocations(cnf.Locations)

	// --run-once: fetch once and exit (useful for cron / manual trigger)
	if *runOnce != "" {
		ctx := context.Background()
		switch *runOnce {
		case "forecasts":
			fetchAll(ctx, domain.TypeForecast, locations, providers, repo, logger)
		case "historical":
			fetchAll(ctx, domain.TypeHistorical, locations, providers, repo, logger)
		default:
			log.Fatalf("--run-once: unknown value %q (must be 'forecasts' or 'historical')", *runOnce)
		}
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Start the background scheduler
	go runScheduler(ctx, cnf, locations, providers, repo, logger)

	// Start the HTTP server
	runHTTP(ctx, cnf, repo, locations, logger)
}

// buildProviders constructs the list of configured weather providers.
func buildProviders(cnf *config.Config, logger *slog.Logger) []portout.Provider {
	httpClient := &http.Client{Timeout: 15 * time.Second}
	var providers []portout.Provider
	if key := cnf.Providers.OpenWeather.APIKey; key != "" {
		providers = append(providers, openweather.New(key, httpClient, logger.With("provider", "openweather")))
	}
	if key := cnf.Providers.AccuWeather.APIKey; key != "" {
		providers = append(providers, accuweather.New(key, httpClient, logger.With("provider", "accuweather")))
	}
	return providers
}

// toDomainLocations converts config locations to domain locations.
func toDomainLocations(locs []config.Location) []domain.Location {
	out := make([]domain.Location, len(locs))
	for i, l := range locs {
		out[i] = domain.Location{
			Name: l.Name,
			Country: domain.Country{
				Name:       l.Country.Name,
				Alpha2Code: l.Country.Alpha2Code,
			},
			Lat:            l.Lat,
			Lon:            l.Lon,
			AccuWeatherKey: l.AccuWeatherKey,
		}
	}
	return out
}

// runScheduler periodically fetches forecast and historical data.
func runScheduler(ctx context.Context, cnf *config.Config, locations []domain.Location, providers []portout.Provider, repo portout.MeasurementRepository, logger *slog.Logger) {
	forecastTicker := time.NewTicker(cnf.ForecastsInterval())
	historicalTicker := time.NewTicker(cnf.HistoricalInterval())
	defer forecastTicker.Stop()
	defer historicalTicker.Stop()

	logger.Info("scheduler: started",
		"forecasts_interval", cnf.ForecastsInterval(),
		"historical_interval", cnf.HistoricalInterval(),
	)

	for {
		select {
		case <-ctx.Done():
			logger.Info("scheduler: shutting down")
			return
		case <-forecastTicker.C:
			fetchAll(ctx, domain.TypeForecast, locations, providers, repo, logger)
		case <-historicalTicker.C:
			fetchAll(ctx, domain.TypeHistorical, locations, providers, repo, logger)
		}
	}
}

// fetchAll fetches measurements of the given type from all providers for all
// locations and persists them.
func fetchAll(ctx context.Context, mType domain.MeasurementType, locations []domain.Location, providers []portout.Provider, repo portout.MeasurementRepository, logger *slog.Logger) {
	for _, loc := range locations {
		for _, p := range providers {
			measurements, err := p.Fetch(ctx, mType, loc)
			if err != nil {
				logger.Error("fetch failed", "type", mType, "location", loc.Name, "err", err)
				continue
			}
			for _, m := range measurements {
				if err := repo.Save(m); err != nil {
					logger.Error("save failed", "type", mType, "location", loc.Name, "err", err)
				}
			}
		}
	}
}

// runHTTP starts the HTTP server and blocks until the context is cancelled.
func runHTTP(ctx context.Context, cnf *config.Config, repo portout.MeasurementRepository, locations []domain.Location, logger *slog.Logger) {
	h := web.NewHandler(repo, locations, cnf.Template.Index, logger.With("component", "http"))

	rh := &regexpRouter{}
	rh.handleFunc(regexp.MustCompile(`^/health$`), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})
	rh.handleFunc(regexp.MustCompile(`/[a-z]{2}/[a-z]{1,}/[0-9]{4}/[0-9]{2}`), h.History)
	rh.handleFunc(regexp.MustCompile(`/`), h.Index)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cnf.Port),
		Handler: rh,
	}

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			logger.Error("http: shutdown error", "err", err)
		}
	}()

	logger.Info("http: listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("http: server error", "err", err)
	}
}

// regexpRouter dispatches requests to handlers by URL pattern.
type regexpRouter struct {
	routes []regexpRoute
}

type regexpRoute struct {
	pattern *regexp.Regexp
	handler http.HandlerFunc
}

func (r *regexpRouter) handleFunc(pattern *regexp.Regexp, handler http.HandlerFunc) {
	r.routes = append(r.routes, regexpRoute{pattern, handler})
}

func (r *regexpRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range r.routes {
		if route.pattern.MatchString(req.URL.Path) {
			route.handler(w, req)
			return
		}
	}
	http.NotFound(w, req)
}
