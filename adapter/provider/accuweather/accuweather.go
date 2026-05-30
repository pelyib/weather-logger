package accuweather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/pelyib/weather-logger/domain"
)

// Provider fetches weather data from the AccuWeather API.
type Provider struct {
	APIKey string
	client *http.Client
	log    *slog.Logger
}

// New creates a Provider with an injected HTTP client and structured logger.
func New(apiKey string, client *http.Client, log *slog.Logger) *Provider {
	return &Provider{APIKey: apiKey, client: client, log: log}
}

func (p *Provider) Fetch(ctx context.Context, mType domain.MeasurementType, loc domain.Location) ([]domain.Measurement, error) {
	switch mType {
	case domain.TypeForecast:
		return p.fetchForecast(ctx, loc)
	case domain.TypeHistorical:
		return p.fetchHistorical(ctx, loc)
	default:
		return nil, fmt.Errorf("accuweather: unknown measurement type %q", mType)
	}
}

// fetchForecast calls the 5-day daily forecast endpoint.
func (p *Provider) fetchForecast(ctx context.Context, loc domain.Location) ([]domain.Measurement, error) {
	if loc.AccuWeatherKey == "" {
		return nil, fmt.Errorf("accuweather forecast: location %q has no AccuWeather key", loc.Name)
	}

	q := url.Values{}
	q.Set("metric", "true")

	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet,
		fmt.Sprintf("https://dataservice.accuweather.com/forecasts/v1/daily/5day/%s", loc.AccuWeatherKey),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("accuweather forecast: build request: %w", err)
	}
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Authorization", "Bearer "+p.APIKey)

	res, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("accuweather forecast: http request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("accuweather forecast: read body: %w", err)
	}

	var decoded struct {
		DailyForecasts []struct {
			Date        string `json:"Date"`
			Temperature struct {
				Minimum struct{ Value float64 } `json:"Minimum"`
				Maximum struct{ Value float64 } `json:"Maximum"`
			} `json:"Temperature"`
		} `json:"DailyForecasts"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, fmt.Errorf("accuweather forecast: parse response: %w", err)
	}

	var results []domain.Measurement
	for _, df := range decoded.DailyForecasts {
		t, err := time.Parse(time.RFC3339, df.Date)
		if err != nil {
			p.log.Warn("accuweather: skip unparseable date", "date", df.Date, "err", err)
			continue
		}
		at, _ := time.Parse("2006-01-02", t.Format("2006-01-02"))
		m, err := domain.NewMeasurement("AccuWeather", domain.TypeForecast,
			df.Temperature.Minimum.Value, df.Temperature.Maximum.Value, at, loc)
		if err != nil {
			p.log.Warn("accuweather: skip invalid forecast measurement", "err", err)
			continue
		}
		results = append(results, m)
	}

	p.log.Info("accuweather: fetched forecast", "location", loc.Name, "count", len(results))
	return results, nil
}

// fetchHistorical calls the 24-hour historical conditions endpoint and
// computes the min/max from hourly readings since midnight today.
func (p *Provider) fetchHistorical(ctx context.Context, loc domain.Location) ([]domain.Measurement, error) {
	if loc.AccuWeatherKey == "" {
		return nil, fmt.Errorf("accuweather historical: location %q has no AccuWeather key", loc.Name)
	}

	q := url.Values{}
	q.Set("apikey", p.APIKey)

	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet,
		fmt.Sprintf("http://dataservice.accuweather.com/currentconditions/v1/%s/historical/24", loc.AccuWeatherKey),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("accuweather historical: build request: %w", err)
	}
	req.URL.RawQuery = q.Encode()

	res, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("accuweather historical: http request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("accuweather historical: read body: %w", err)
	}

	var decoded []struct {
		EpochTime   int64 `json:"EpochTime"`
		Temperature struct {
			Metric struct{ Value float64 } `json:"Metric"`
		} `json:"Temperature"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, fmt.Errorf("accuweather historical: parse response: %w", err)
	}

	today, _ := time.Parse("2006-01-02", time.Now().UTC().Format("2006-01-02"))
	todayUnix := today.Unix()
	yesterday := today.Add(-24 * time.Hour)

	var min, max float64 = 60, -55
	for _, obs := range decoded {
		if obs.EpochTime < todayUnix {
			continue
		}
		v := obs.Temperature.Metric.Value
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	m, err := domain.NewMeasurement("AccuWeather", domain.TypeHistorical, min, max, yesterday, loc)
	if err != nil {
		return nil, fmt.Errorf("accuweather historical: %w", err)
	}

	p.log.Info("accuweather: fetched historical", "location", loc.Name, "date", yesterday.Format("2006-01-02"))
	return []domain.Measurement{m}, nil
}
