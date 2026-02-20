package openweather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pelyib/weather-logger/domain"
)

// Provider fetches weather data from the OpenWeatherMap API.
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
		return nil, fmt.Errorf("openweather: unknown measurement type %q", mType)
	}
}

// fetchForecast calls the One Call API and returns the daily forecast.
func (p *Provider) fetchForecast(ctx context.Context, loc domain.Location) ([]domain.Measurement, error) {
	q := url.Values{}
	q.Set("lat", fmt.Sprintf("%f", loc.Lat))
	q.Set("lon", fmt.Sprintf("%f", loc.Lon))
	q.Set("exclude", "current,minutely,hourly,alerts")
	q.Set("appid", p.APIKey)
	q.Set("units", "metric")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.openweathermap.org/data/2.5/onecall", nil)
	if err != nil {
		return nil, fmt.Errorf("openweather forecast: build request: %w", err)
	}
	req.URL.RawQuery = q.Encode()

	res, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openweather forecast: http request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("openweather forecast: read body: %w", err)
	}

	var decoded struct {
		Daily []struct {
			Dt   int64 `json:"dt"`
			Temp struct {
				Min float64 `json:"min"`
				Max float64 `json:"max"`
			} `json:"temp"`
		} `json:"daily"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, fmt.Errorf("openweather forecast: parse response: %w", err)
	}

	var results []domain.Measurement
	for _, df := range decoded.Daily {
		at, _ := time.Parse("2006-01-02", time.Unix(df.Dt, 0).UTC().Format("2006-01-02"))
		m, err := domain.NewMeasurement("OpenWeather", domain.TypeForecast, df.Temp.Min, df.Temp.Max, at, loc)
		if err != nil {
			p.log.Warn("openweather: skip invalid forecast measurement", "err", err)
			continue
		}
		results = append(results, m)
	}

	p.log.Info("openweather: fetched forecast", "location", loc.Name, "count", len(results))
	return results, nil
}

// fetchHistorical calls the One Call timemachine endpoint for yesterday.
func (p *Provider) fetchHistorical(ctx context.Context, loc domain.Location) ([]domain.Measurement, error) {
	today, _ := time.Parse("2006-01-02", time.Now().UTC().Format("2006-01-02"))
	yesterday := today.Add(-24 * time.Hour)

	q := url.Values{}
	q.Set("lat", fmt.Sprintf("%f", loc.Lat))
	q.Set("lon", fmt.Sprintf("%f", loc.Lon))
	q.Set("appid", p.APIKey)
	q.Set("units", "metric")
	q.Set("dt", strconv.FormatInt(yesterday.Unix(), 10))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.openweathermap.org/data/2.5/onecall/timemachine", nil)
	if err != nil {
		return nil, fmt.Errorf("openweather historical: build request: %w", err)
	}
	req.URL.RawQuery = q.Encode()

	res, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openweather historical: http request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("openweather historical: read body: %w", err)
	}

	var decoded struct {
		Hourly []struct {
			Dt   int64   `json:"dt"`
			Temp float64 `json:"temp"`
		} `json:"hourly"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, fmt.Errorf("openweather historical: parse response: %w", err)
	}

	var min, max float64 = 60, -60
	todayUnix := today.Unix()
	yesterdayUnix := yesterday.Unix()

	for _, h := range decoded.Hourly {
		if h.Dt < yesterdayUnix || h.Dt >= todayUnix {
			continue
		}
		if h.Temp < min {
			min = h.Temp
		}
		if h.Temp > max {
			max = h.Temp
		}
	}

	m, err := domain.NewMeasurement("OpenWeather", domain.TypeHistorical, min, max, yesterday, loc)
	if err != nil {
		return nil, fmt.Errorf("openweather historical: %w", err)
	}

	p.log.Info("openweather: fetched historical", "location", loc.Name, "date", yesterday.Format("2006-01-02"))
	return []domain.Measurement{m}, nil
}
