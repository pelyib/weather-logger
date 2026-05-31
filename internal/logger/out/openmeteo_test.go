package out

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func stubClient(body string) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}
}

// omFixture mirrors the actual API response from 2026-05-31, converted to
// timeformat=unixtime. Timestamps: 2026-05-31T00:00:00Z = 1780185600, +86400/day.
//
// ecmwf_ifs:    7 non-null days (full forecast horizon)
// gem_regional: 3 non-null + 4 null  (shorter horizon)
// kma_seamless: all null              (unavailable)
const omFixture = `{
	"latitude": 49.736378,
	"longitude": 6.5511265,
	"utc_offset_seconds": 7200,
	"timezone": "Europe/Berlin",
	"daily": {
		"time": [1780185600, 1780272000, 1780358400, 1780444800, 1780531200, 1780617600, 1780704000],
		"temperature_2m_max_ecmwf_ifs": [25.2, 23.1, 20.7, 19.1, 18.0, 19.4, 19.0],
		"temperature_2m_min_ecmwf_ifs": [16.4, 12.1, 13.5, 12.3, 13.0, 12.4, 11.9],
		"temperature_2m_max_gem_regional": [24.6, 22.1, 23.2, null, null, null, null],
		"temperature_2m_min_gem_regional": [16.7, 13.1, 13.6, null, null, null, null],
		"temperature_2m_max_kma_seamless": [null, null, null, null, null, null, null],
		"temperature_2m_min_kma_seamless": [null, null, null, null, null, null, null]
	}
}`

func makeTestProvider(models []string, client *http.Client) omForecast {
	cnf := &shared.LoggerCnf{}
	cnf.ForecastProviders.OpenMeteo.Models = models
	cnf.ForecastProviders.OpenMeteo.Timezone = "Europe/Berlin"
	return omForecast{cnf: cnf, l: shared.MakeNullLogger(), client: client}
}

func TestOMForecast_GetMeasurement_ResultCount(t *testing.T) {
	tests := []struct {
		name   string
		models []string
		want   int
	}{
		{"ecmwf_ifs full horizon", []string{"ecmwf_ifs"}, 7},
		{"gem_regional partial horizon", []string{"gem_regional"}, 3},
		{"kma_seamless all null", []string{"kma_seamless"}, 0},
		{"unknown model not in response", []string{"nonexistent_model"}, 0},
		{"no models configured", []string{}, 0},
		{"multiple models combined", []string{"ecmwf_ifs", "gem_regional", "kma_seamless"}, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := makeTestProvider(tt.models, stubClient(omFixture))
			got := p.GetMeasurement(shared.SearchRequest{})
			if len(got) != tt.want {
				t.Errorf("got %d results, want %d", len(got), tt.want)
			}
		})
	}
}

func TestOMForecast_GetMeasurement_SourcePerModel(t *testing.T) {
	p := makeTestProvider([]string{"ecmwf_ifs"}, stubClient(omFixture))
	results := p.GetMeasurement(shared.SearchRequest{})

	for _, r := range results {
		if r.Source != "OpenMeteo/ecmwf_ifs" {
			t.Errorf("Source: got %q, want %q", r.Source, "OpenMeteo/ecmwf_ifs")
		}
	}
}

func TestOMForecast_GetMeasurement_TypeIsForecast(t *testing.T) {
	p := makeTestProvider([]string{"ecmwf_ifs"}, stubClient(omFixture))
	results := p.GetMeasurement(shared.SearchRequest{})

	for _, r := range results {
		if r.Type != shared.MeasurementResult_Type_Forecast {
			t.Errorf("Type: got %q, want %q", r.Type, shared.MeasurementResult_Type_Forecast)
		}
	}
}

func TestOMForecast_GetMeasurement_FirstDayMinMax(t *testing.T) {
	p := makeTestProvider([]string{"ecmwf_ifs"}, stubClient(omFixture))
	results := p.GetMeasurement(shared.SearchRequest{})

	if len(results) == 0 {
		t.Fatal("expected results, got none")
	}

	first := results[0]
	if first.Min != 16.4 {
		t.Errorf("Min: got %v, want 16.4", first.Min)
	}
	if first.Max != 25.2 {
		t.Errorf("Max: got %v, want 25.2", first.Max)
	}
}

func TestOMForecast_GetMeasurement_AtMatchesTimestamp(t *testing.T) {
	p := makeTestProvider([]string{"ecmwf_ifs"}, stubClient(omFixture))
	results := p.GetMeasurement(shared.SearchRequest{})

	if len(results) == 0 {
		t.Fatal("expected results, got none")
	}

	wantAt := time.Unix(1780185600, 0).UTC().Format(time.RFC3339)
	if results[0].At != wantAt {
		t.Errorf("At: got %q, want %q", results[0].At, wantAt)
	}
}

// omHistoricalFixture mirrors the example API response provided in the task spec.
// Timestamp 1780092000 = 2026-05-30T22:00:00Z (midnight Europe/Berlin CEST).
const omHistoricalFixture = `{
	"latitude": 52.54833,
	"longitude": 13.407822,
	"utc_offset_seconds": 7200,
	"timezone": "Europe/Berlin",
	"daily_units": {"time": "unixtime", "temperature_2m_max": "°C", "temperature_2m_min": "°C"},
	"daily": {
		"time": [1780092000],
		"temperature_2m_max": [22.1],
		"temperature_2m_min": [17.9]
	}
}`

func makeTestHistoricalProvider(client *http.Client) omHistorical {
	cnf := &shared.LoggerCnf{}
	cnf.ForecastProviders.OpenMeteo.Timezone = "Europe/Berlin"
	return omHistorical{cnf: cnf, l: shared.MakeNullLogger(), client: client}
}

func TestOMHistorical_GetMeasurement_HappyPath(t *testing.T) {
	p := makeTestHistoricalProvider(stubClient(omHistoricalFixture))
	results := p.GetMeasurement(shared.SearchRequest{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.Min != 17.9 {
		t.Errorf("Min: got %v, want 17.9", r.Min)
	}
	if r.Max != 22.1 {
		t.Errorf("Max: got %v, want 22.1", r.Max)
	}
}

func TestOMHistorical_GetMeasurement_Source(t *testing.T) {
	p := makeTestHistoricalProvider(stubClient(omHistoricalFixture))
	results := p.GetMeasurement(shared.SearchRequest{})
	if len(results) == 0 {
		t.Fatal("expected results, got none")
	}
	if results[0].Source != "OpenMeteo" {
		t.Errorf("Source: got %q, want %q", results[0].Source, "OpenMeteo")
	}
}

func TestOMHistorical_GetMeasurement_TypeIsHistorical(t *testing.T) {
	p := makeTestHistoricalProvider(stubClient(omHistoricalFixture))
	results := p.GetMeasurement(shared.SearchRequest{})
	if len(results) == 0 {
		t.Fatal("expected results, got none")
	}
	if results[0].Type != shared.MeasurementResult_Type_Historical {
		t.Errorf("Type: got %q, want %q", results[0].Type, shared.MeasurementResult_Type_Historical)
	}
}

func TestOMHistorical_GetMeasurement_AtMatchesTimestamp(t *testing.T) {
	p := makeTestHistoricalProvider(stubClient(omHistoricalFixture))
	results := p.GetMeasurement(shared.SearchRequest{})
	if len(results) == 0 {
		t.Fatal("expected results, got none")
	}
	want := time.Unix(1780092000, 0).UTC().Format(time.RFC3339)
	if results[0].At != want {
		t.Errorf("At: got %q, want %q", results[0].At, want)
	}
}

func TestOMHistorical_GetMeasurement_LocationPropagated(t *testing.T) {
	loc := shared.Location{Name: "Berlin"}
	p := makeTestHistoricalProvider(stubClient(omHistoricalFixture))
	results := p.GetMeasurement(shared.SearchRequest{Loc: loc})
	if len(results) == 0 {
		t.Fatal("expected results, got none")
	}
	if results[0].Loc.Name != "Berlin" {
		t.Errorf("Loc.Name: got %q, want %q", results[0].Loc.Name, "Berlin")
	}
}

func TestOMHistorical_GetMeasurement_EmptyDailyArrays(t *testing.T) {
	const fixture = `{"daily":{"time":[],"temperature_2m_max":[],"temperature_2m_min":[]}}`
	p := makeTestHistoricalProvider(stubClient(fixture))
	results := p.GetMeasurement(shared.SearchRequest{})
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty daily arrays, got %d", len(results))
	}
}

func TestOMHistorical_GetMeasurement_MalformedJSON(t *testing.T) {
	p := makeTestHistoricalProvider(stubClient("not json"))
	results := p.GetMeasurement(shared.SearchRequest{})
	if len(results) != 0 {
		t.Errorf("expected 0 results for malformed JSON, got %d", len(results))
	}
}

func TestOMHistorical_GetMeasurement_HTTPError(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("connection refused")
		}),
	}
	p := makeTestHistoricalProvider(client)
	results := p.GetMeasurement(shared.SearchRequest{})
	if len(results) != 0 {
		t.Errorf("expected 0 results on HTTP error, got %d", len(results))
	}
}

func TestOMForecast_GetMeasurement_LocationPropagated(t *testing.T) {
	loc := shared.Location{Name: "Luxembourg", Country: struct {
		Name       string `yaml:"name" json:"name"`
		Alpha2Code string `yaml:"alpha2Code" json:"alpha2Code"`
	}{Alpha2Code: "LU"}}

	p := makeTestProvider([]string{"ecmwf_ifs"}, stubClient(omFixture))
	results := p.GetMeasurement(shared.SearchRequest{Loc: loc})

	for _, r := range results {
		if r.Loc.Name != loc.Name {
			t.Errorf("Loc.Name: got %q, want %q", r.Loc.Name, loc.Name)
		}
	}
}
