package out

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

// =====================
// ===== FORECASTS =====
// =====================

func TestAccuweatherForecast_sourceId_returnsIt(t *testing.T) {
	awf := awForecast{}
	if awf.sourceId() != "accuweather.forecast" {
		t.Errorf("Expected accuweather.forecast, got %s", awf.sourceId())
	}
}

func TestAccuweatherForecast_fetch_returnsError_whenCallFailed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
	}))
	defer ts.Close()

	awf := awForecast{cnf: &shared.LoggerCnf{
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
			AccuWeather: struct {
				Host  string
				AppId string
			}{
				Host:  ts.URL,
				AppId: "app-id",
			}},
	}}

	result, err := awf.fetch(shared.SearchRequest{
		Loc: shared.Location{
			Providers: struct {
				AccuWeather struct {
					Locationkey string `yaml:"locationKey" json:"locationKey"`
				} `yaml:"accuWeather" json:"accuWeather"`
			}{
				AccuWeather: struct {
					Locationkey string `yaml:"locationKey" json:"locationKey"`
				}{
					Locationkey: "location-key",
				},
			},
		},
	})

	if err == nil {
		t.Error("Expected error, got nothing")
	}

	if result != nil {
		t.Errorf("Expected nil as result, got %v", result)
	}

	if err.Error() != "Fetching Forecasts from Accuweather failed, HTTP status code: 400 Bad Request" {
		t.Errorf("Expected error message mismatch, got %s", err.Error())
	}
}

func TestAccuweatherForecast_fetch_returnsRawResponse_whenCallSucceeds(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appId := r.URL.Query().Get("apikey")
		if appId != "app-id" {
			w.WriteHeader(401)
			t.Errorf("Expected apikey mismatch, got %s", appId)

			return
		}
		metric := r.URL.Query().Get("metric")
		if metric != "true" {
			w.WriteHeader(400)
			t.Errorf("Expected metric mismatch, got %s", metric)

			return
		}

		path := r.URL.Path
		if path != "/forecasts/v1/daily/5day/location-key" {
			w.WriteHeader(404)
			t.Errorf("Expected path mismatch, got %s", path)

			return
		}

		w.WriteHeader(200)
		w.Write([]byte("API raw response"))
	}))
	defer ts.Close()

	awf := awForecast{cnf: &shared.LoggerCnf{
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
			AccuWeather: struct {
				Host  string
				AppId string
			}{
				Host:  ts.URL,
				AppId: "app-id",
			}},
	}}

	result, err := awf.fetch(shared.SearchRequest{
		Loc: shared.Location{
			Providers: struct {
				AccuWeather struct {
					Locationkey string `yaml:"locationKey" json:"locationKey"`
				} `yaml:"accuWeather" json:"accuWeather"`
			}{
				AccuWeather: struct {
					Locationkey string `yaml:"locationKey" json:"locationKey"`
				}{
					Locationkey: "location-key",
				},
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

func TestAccuweatherForecast_mapToMeasurement_returnsACollection_whenRawIsValid(t *testing.T) {
	awf := awForecast{
		now: func() time.Time {
			return time.Date(2024, 12, 27, 10, 11, 12, 0, time.UTC)
		},
	}

	data, err := os.ReadFile("./../../../testdata/out/accuweather_forecast.json")

	if err != nil {
		t.Errorf("Tried to load testdata, but got error: %s", err.Error())
	}

	result, err := awf.mapToMeasurements(data, shared.Location{})

	if err != nil {
		t.Errorf("Expected no error, got %s", err.Error())
	}

	if len(result) != 5 {
		t.Errorf("Expected exactly 5 items, got %v", len(result))
	}

	minMaxValues := []map[string]float32{
		{"min": 1.4, "max": 12.0},
		{"min": 2.6, "max": 4.5},
		{"min": 6.6, "max": 8.4},
		{"min": 6.1, "max": 9.1},
		{"min": 6.4, "max": 8.5},
	}
	atValues := []string{
		"2024-12-27T00:00:00Z",
		"2024-12-28T00:00:00Z",
		"2024-12-29T00:00:00Z",
		"2024-12-30T00:00:00Z",
		"2024-12-31T00:00:00Z",
	}

	for i, item := range result {
		if item.Source != "AccuWeather" {
			t.Errorf("Expected source mismatch, got %s", item.Source)
		}

		if item.Type != shared.MeasurementResult_Type_Forecast {
			t.Errorf("Expected type mismatch, got %s", item.Type)
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

		if item.RecordedAt != "2024-12-27T10:11:12Z" {
			t.Errorf("Expected recordedAt mismatch, got %s", item.RecordedAt)
		}
	}
}

// ======================
// ===== HISTORICAL =====
// ======================

func TesrAccuweatherHistorical_sourceId_returnsIt(t *testing.T) {
	awh := awHistorical{}
	if awh.sourceId() != "accuweather.historical" {
		t.Errorf("Expected accuweather.historical, got %s", awh.sourceId())
	}
}

func TestAccuweatherHistorical_fetch_returnsError_whenCallFailed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()
	awh := awHistorical{cnf: &shared.LoggerCnf{
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
			AccuWeather: struct {
				Host  string
				AppId string
			}{
				Host:  ts.URL,
				AppId: "app-id",
			}},
	}}

	result, err := awh.fetch(shared.SearchRequest{
		Loc: shared.Location{
			Providers: struct {
				AccuWeather struct {
					Locationkey string `yaml:"locationKey" json:"locationKey"`
				} `yaml:"accuWeather" json:"accuWeather"`
			}{
				AccuWeather: struct {
					Locationkey string `yaml:"locationKey" json:"locationKey"`
				}{
					Locationkey: "location-key",
				},
			},
		},
	})

	if err == nil {
		t.Error("Expected error, got nothing")
	}

	if err.Error() != "Fetching Historical from Accuweather failed, HTTP status code: 500 Internal Server Error" {
		t.Errorf("Expected error message mismatch, got %s", err.Error())
	}

	if result != nil {
		t.Errorf("Expected nil as result, got %v", result)
	}
}

func TestAccuweatherHistorical_fetch_returnsRawResponse_whenCallSucceeds(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appId := r.URL.Query().Get("apikey")
		if appId != "app-id" {
			w.WriteHeader(401)
			t.Errorf("Expected apikey mismatch, got %s", appId)

			return
		}

		path := r.URL.Path
		if path != "/currentconditions/v1/location-key/historical/24" {
			w.WriteHeader(404)
			t.Errorf("Expected path mismatch, got %s", path)

			return
		}

		w.WriteHeader(200)
		w.Write([]byte("API raw response"))
	}))
	defer ts.Close()
	awh := awHistorical{cnf: &shared.LoggerCnf{
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
			AccuWeather: struct {
				Host  string
				AppId string
			}{
				Host:  ts.URL,
				AppId: "app-id",
			}},
	}}

	result, err := awh.fetch(shared.SearchRequest{
		Loc: shared.Location{
			Providers: struct {
				AccuWeather struct {
					Locationkey string `yaml:"locationKey" json:"locationKey"`
				} `yaml:"accuWeather" json:"accuWeather"`
			}{
				AccuWeather: struct {
					Locationkey string `yaml:"locationKey" json:"locationKey"`
				}{
					Locationkey: "location-key",
				},
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

func TestAccuweatherHistorical_mapToMeasurement_returnsACollection_whenRawIsValid(t *testing.T) {
	awh := awHistorical{
		now: func() time.Time {
			return time.Date(2024, 12, 27, 10, 11, 12, 0, time.UTC)
		},
	}

	data, err := os.ReadFile("./../../../testdata/out/accuweather_historical.json")

	if err != nil {
		t.Errorf("Tried to load testdata, but got error: %s", err.Error())
	}

	result, err := awh.mapToMeasurements(data, shared.Location{})
	if err != nil {
		t.Errorf("Expected no error, got %s", err.Error())
	}

	if len(result) != 1 {
		t.Errorf("Expected exactly 1 item, got %v", len(result))
	}

	if result[0].Source != "AccuWeather" {
		t.Errorf("Expected source mismatch, got %s", result[0].Source)
	}

	if result[0].Type != shared.MeasurementResult_Type_Historical {
		t.Errorf("Expected type mismatch, got %s", result[0].Type)
	}

	if result[0].At != "2024-12-26T00:00:00Z" {
		t.Errorf("Expected at mismatch, got %s", result[0].At)
	}

	if result[0].RecordedAt != "2024-12-27T10:11:12Z" {
		t.Errorf("Expected recordedAt mismatch, got %s", result[0].RecordedAt)
	}

	if result[0].Min != 2.2 {
		t.Errorf("Expected min mismatch, got %f", result[0].Min)
	}

	if result[0].Max != 5 {
		t.Errorf("Expected max mismatch, got %f", result[0].Max)
	}
}
