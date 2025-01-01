package out

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

// =====================
// ===== FORECASTS =====
// =====================

func TestOpenWeatherForecast_sourceId_returnsIt(t *testing.T) {
	awf := owForecast{}
	if awf.sourceId() != "openweather.forecast" {
		t.Errorf("Expected openweather.forecast, got %s", awf.sourceId())
	}
}

func TestOpenWeatherForecast_fetch_returnsError_whenCallFailed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
	}))
	defer ts.Close()

	sut := owForecast{cnf: &shared.LoggerCnf{
		ForecastProviders: struct {
			OpenWeather struct {
				Host  string
				AppId string
			} `yaml:"openWeather"`
			AccuWeather struct {
				Host  string
				AppId string
			} `yaml:"accuweather"`
		}{
			OpenWeather: struct {
				Host  string
				AppId string
			}{
				Host:  ts.URL,
				AppId: "app-id",
			}},
	}}

	result, err := sut.fetch(shared.SearchRequest{
		Loc: shared.Location{
			GeoLocation: struct {
				Langitude float64 `yaml:"langitude" json:"langitude"`
				Longitude float64 `yaml:"longitude" json:"longitude"`
			}{
				Langitude: 1.0,
				Longitude: 1.0,
			},
		},
	})

	if err == nil {
		t.Error("Expected error, got nothing")
	}

	if result != nil {
		t.Errorf("Expected nil as result, got %v", result)
	}

	if err.Error() != "Fetching forecasts from OpenWeather failed, HTTP status code: 400 Bad Request" {
		t.Errorf("Expected error message mismatch, got %s", err.Error())
	}
}

func TestOpenWeatherForecast_fetch_returnsRawResponse_whenCallSucceeds(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path != "/data/3.0/onecall" {
			w.WriteHeader(404)
			t.Errorf("Expected path mismatch, got %s", path)

			return
		}

		appId := r.URL.Query().Get("appid")
		if appId != "app-id" {
			w.WriteHeader(401)
			t.Errorf("Expected apikey mismatch, got %s", appId)

			return
		}

		lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
		if lat != 13.0 {
			w.WriteHeader(404)
			t.Errorf("Expected langitude mismatch, got %f", lat)

			return
		}

		lon, _ := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
		if lon != 42.0 {
			w.WriteHeader(404)
			t.Errorf("Expected longitude mismatch, got %f", lon)

			return
		}

		units := r.URL.Query().Get("units")
		if units != "metric" {
			w.WriteHeader(404)
			t.Errorf("Expected units mismatch, got %s", units)

			return
		}

		exclude := r.URL.Query().Get("exclude")
		if exclude != "current,minutely,hourly,alerts" {
			w.WriteHeader(404)
			t.Errorf("Expected exclude mismatch, got %s", exclude)

			return
		}

		w.WriteHeader(200)
		w.Write([]byte("API raw response"))
	}))
	defer ts.Close()

	sut := owForecast{cnf: &shared.LoggerCnf{
		ForecastProviders: struct {
			OpenWeather struct {
				Host  string
				AppId string
			} `yaml:"openWeather"`
			AccuWeather struct {
				Host  string
				AppId string
			} `yaml:"accuweather"`
		}{
			OpenWeather: struct {
				Host  string
				AppId string
			}{
				Host:  ts.URL,
				AppId: "app-id",
			}},
	}}

	result, err := sut.fetch(shared.SearchRequest{
		Loc: shared.Location{
			GeoLocation: struct {
				Langitude float64 `yaml:"langitude" json:"langitude"`
				Longitude float64 `yaml:"longitude" json:"longitude"`
			}{
				Langitude: 13.0,
				Longitude: 42.0,
			},
		},
	})

	if err != nil {
		t.Errorf("Expected no error, got %s", err.Error())
	}

	if result == nil {
		t.Errorf("Expected []byte, got %v", result)
	}

	if string(result) != "API raw response" {
		t.Errorf("Expected response mismatch, got %s", string(result))
	}
}

func TestOpenWeatherForecast_mapToMeasurement_returnsError_whenRawIsInvalid(t *testing.T) {
	sut := owForecast{}
	measurements, err := sut.mapToMeasurements([]byte("invalid"), shared.Location{})

	if err == nil {
		t.Error("Expected error, got nothing")
	}

	if err.Error() != "Could not parse raw response body | Reason: invalid character 'i' looking for beginning of value" {
		t.Errorf("Expected error message mismatch, got %s", err.Error())
	}

	if measurements != nil {
		t.Errorf("Expected nil as measurements, got %v", measurements)
	}
}

func TestOpenWeatherForecast_mapToMeasurement_returnsACollection_whenRawIsValid(t *testing.T) {
	data, err := os.ReadFile("./../../../testdata/out/openweather_forecast.json")

	if err != nil {
		t.Errorf("Tried to load testdata, but got error: %s", err.Error())
	}

	sut := owForecast{
		now: func() time.Time {
			return time.Date(2025, 01, 01, 10, 11, 12, 0, time.UTC)
		},
	}
	measurements, err := sut.mapToMeasurements(data, shared.Location{})

	if err != nil {
		t.Errorf("Expected no error, got %s", err.Error())
	}

	if len(measurements) != 8 {
		t.Errorf("Expected 8 elements, got %d", len(measurements))
	}

	minMaxValues := []map[string]float32{
		{"min": -2.71, "max": 3.62},
		{"min": 1.06, "max": 9.27},
		{"min": 1.08, "max": 5.98},
		{"min": -0.17, "max": 2.7},
		{"min": -0.83, "max": 5.86},
		{"min": 6.34, "max": 10.88},
		{"min": 7.28, "max": 13.2},
		{"min": 6.27, "max": 11.84},
	}
	atValues := []string{
		"2025-01-01T00:00:00Z",
		"2025-01-02T00:00:00Z",
		"2025-01-03T00:00:00Z",
		"2025-01-04T00:00:00Z",
		"2025-01-05T00:00:00Z",
		"2025-01-06T00:00:00Z",
		"2025-01-07T00:00:00Z",
		"2025-01-08T00:00:00Z",
	}

	for i, item := range measurements {
		if item.Source != "OpenWeather" {
			t.Errorf("Expected OpenWeather, got %s", item.Source)
		}

		if item.Type != shared.MeasurementResult_Type_Forecast {
			t.Errorf("Expected Forecast, got %s", item.Type)
		}

		if item.Min != minMaxValues[i]["min"] {
			t.Errorf("Expected min mismatch, got %f", item.Min)
		}

		if item.Max != minMaxValues[i]["max"] {
			t.Errorf("Expected max mismatch, got %f", item.Max)
		}

		if item.At != atValues[i] {
			t.Errorf("Expected at mismatch, got %s", item.At)
		}

		if item.RecordedAt != "2025-01-01T10:11:12Z" {
			t.Errorf("Expected recordedAt mismatch, got %s", item.RecordedAt)
		}
	}
}

// ======================
// ===== HISTORICAL =====
// ======================

func TestOpenWeatherHistorical_sourceId_returnsIt(t *testing.T) {
	owh := owHistorical{}
	if owh.sourceId() != "openweather.historical" {
		t.Errorf("Expected openweather.historical, got %s", owh.sourceId())
	}
}

func TestOpenWeatherHistorical_fetch_returnsError_whenCallFailed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
	}))
	defer ts.Close()

	sut := owHistorical{
		now: func() time.Time {
			return time.Date(2025, 01, 01, 10, 11, 12, 0, time.UTC)
		},
		cnf: &shared.LoggerCnf{
			ForecastProviders: struct {
				OpenWeather struct {
					Host  string
					AppId string
				} `yaml:"openWeather"`
				AccuWeather struct {
					Host  string
					AppId string
				} `yaml:"accuweather"`
			}{
				OpenWeather: struct {
					Host  string
					AppId string
				}{
					Host:  ts.URL,
					AppId: "app-id",
				}},
		}}

	result, err := sut.fetch(shared.SearchRequest{
		Loc: shared.Location{
			GeoLocation: struct {
				Langitude float64 `yaml:"langitude" json:"langitude"`
				Longitude float64 `yaml:"longitude" json:"longitude"`
			}{
				Langitude: 13.0,
				Longitude: 42.0,
			},
		},
	})

	if err == nil {
		t.Error("Expected error, got nothing")
	}

	if result != nil {
		t.Errorf("Expected nil as result, got %v", result)
	}
}

func TestOpenWeatherHistorical_fetch_returnsRawResponse_whenCallSucceeds(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/3.0/onecall/day_summary" {
			w.WriteHeader(404)
			t.Errorf("Expected path mismatch, got %s", r.URL.Path)

			return
		}

		appId := r.URL.Query().Get("appid")
		if appId != "app-id" {
			w.WriteHeader(401)
			t.Errorf("Expected appid mismatch, got %s", appId)

			return
		}

		lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
		if lat != 13.0 {
			w.WriteHeader(404)
			t.Errorf("Expected langitude mismatch, got %f", lat)

			return
		}

		lon, _ := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
		if lon != 42.0 {
			w.WriteHeader(404)
			t.Errorf("Expected longitude mismatch, got %f", lon)

			return
		}

		units := r.URL.Query().Get("units")
		if units != "metric" {
			w.WriteHeader(404)
			t.Errorf("Expected units mismatch, got %s", units)

			return
		}

		date := r.URL.Query().Get("date")
		if date != "2024-12-31" {
			w.WriteHeader(404)
			t.Errorf("Expected date mismatch, got %s", date)

			return
		}

		w.WriteHeader(200)
		w.Write([]byte("API raw response"))
	}))
	defer ts.Close()

	sut := owHistorical{
		now: func() time.Time {
			return time.Date(2025, 01, 01, 10, 11, 12, 0, time.UTC)
		},
		cnf: &shared.LoggerCnf{
			ForecastProviders: struct {
				OpenWeather struct {
					Host  string
					AppId string
				} `yaml:"openWeather"`
				AccuWeather struct {
					Host  string
					AppId string
				} `yaml:"accuweather"`
			}{
				OpenWeather: struct {
					Host  string
					AppId string
				}{
					Host:  ts.URL,
					AppId: "app-id",
				}},
		}}

	result, err := sut.fetch(shared.SearchRequest{
		Loc: shared.Location{
			GeoLocation: struct {
				Langitude float64 `yaml:"langitude" json:"langitude"`
				Longitude float64 `yaml:"longitude" json:"longitude"`
			}{
				Langitude: 13.0,
				Longitude: 42.0,
			},
		},
	})

	if err != nil {
		t.Errorf("Expected no error, got %s", err.Error())

	}

	if result == nil {
		t.Errorf("Expected []byte, got %v", result)
	}

	if string(result) != "API raw response" {
		t.Errorf("Expected response mismatch, got %s", string(result))
	}
}

func TestOpenWeatherHistorical_mapToMeasurement_returnsError_whenRawIsInvalid(t *testing.T) {
	sut := owHistorical{}
	measurements, err := sut.mapToMeasurements([]byte("invalid"), shared.Location{})

	if err == nil {
		t.Error("Expected error, got nothing")
	}

	if err.Error() != "Could not parse raw response body | Reason: invalid character 'i' looking for beginning of value" {
		t.Errorf("Expected error message mismatch, got %s", err.Error())
	}

	if measurements != nil {
		t.Errorf("Expected nil as measurements, got %v", measurements)
	}
}

func TestOpenWeatherHistorical_mapToMeasurement_returnsACollection_whenRawIsValid(t *testing.T) {
	data, err := os.ReadFile("./../../../testdata/out/openweather_day_summary.json")

	if err != nil {
		t.Errorf("Tried to load testdata, but got error: %s", err.Error())
	}

	sut := owHistorical{
		now: func() time.Time {
			return time.Date(2025, 01, 01, 10, 11, 12, 0, time.UTC)
		},
	}
	measurements, err := sut.mapToMeasurements(data, shared.Location{})

	if err != nil {
		t.Errorf("Expected no error, got %s", err.Error())
	}

	if len(measurements) != 1 {
		t.Errorf("Expected 1 elements, got %d", len(measurements))
	}

	item := measurements[0]
	if item.Source != "OpenWeather" {
		t.Errorf("Expected OpenWeather, got %s", item.Source)
	}

	if item.Type != shared.MeasurementResult_Type_Historical {
		t.Errorf("Expected Historical, got %s", item.Type)
	}

	if item.Min != -2.71 {
		t.Errorf("Expected min mismatch, got %f", item.Min)
	}

	if item.Max != -0.92 {
		t.Errorf("Expected max mismatch, got %f", item.Max)
	}

	if item.At != "2024-12-31T00:00:00Z" {
		t.Errorf("Expected at mismatch, got %s", item.At)
	}

	if item.RecordedAt != "2025-01-01T10:11:12Z" {
		t.Errorf("Expected recordedAt mismatch, got %s", item.RecordedAt)
	}
}
