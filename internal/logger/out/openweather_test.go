package out

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/pelyib/weather-logger/internal/shared"
)

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
