package out

import (
	"testing"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

// owFixture has entries spanning 3 UTC days:
//
//	day 0 (1780185600 = 2026-05-31T00:00:00Z): 2 entries — min=10.0/11.0, max=20.0/21.0
//	day 1 (1780272000 = 2026-06-01T00:00:00Z): 1 entry
//	day 2 (1780358400 = 2026-06-02T00:00:00Z): 1 entry
const owFixture = `{
	"cod": "200",
	"list": [
		{"dt": 1780185600, "main": {"temp_min": 10.0, "temp_max": 20.0}},
		{"dt": 1780196400, "main": {"temp_min": 11.0, "temp_max": 21.0}},
		{"dt": 1780272000, "main": {"temp_min": 12.0, "temp_max": 22.0}},
		{"dt": 1780358400, "main": {"temp_min":  9.0, "temp_max": 19.0}}
	]
}`

func makeOWForecastProvider(client interface{ Do(*interface{}) error }, db interface{}) owForecast {
	// helper not used directly; tests construct owForecast inline for clarity
	panic("use inline construction")
}

func newOWForecast(client interface{}, db interface{}) owForecast {
	panic("unused — see tests below")
}

func TestOWForecast_GetMeasurement_ResultCountByDay(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		want    int
	}{
		{"3 distinct days", owFixture, 3},
		{"empty list", `{"list":[]}`, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cnf := &shared.LoggerCnf{}
			p := owForecast{cnf: cnf, l: shared.MakeNullLogger(), client: stubClient(tt.fixture), db: nil}
			got := p.GetMeasurement(shared.SearchRequest{})
			if len(got) != tt.want {
				t.Errorf("got %d results, want %d", len(got), tt.want)
			}
		})
	}
}

func TestOWForecast_GetMeasurement_DayAggregation(t *testing.T) {
	cnf := &shared.LoggerCnf{}
	p := owForecast{cnf: cnf, l: shared.MakeNullLogger(), client: stubClient(owFixture), db: nil}
	results := p.GetMeasurement(shared.SearchRequest{})

	if len(results) == 0 {
		t.Fatal("expected results, got none")
	}
	first := results[0]
	// day 0 has two entries: min should be lowest (10.0), max should be highest (21.0)
	if first.Min != 10.0 {
		t.Errorf("Min: got %v, want 10.0", first.Min)
	}
	if first.Max != 21.0 {
		t.Errorf("Max: got %v, want 21.0", first.Max)
	}
}

func TestOWForecast_GetMeasurement_Source(t *testing.T) {
	cnf := &shared.LoggerCnf{}
	p := owForecast{cnf: cnf, l: shared.MakeNullLogger(), client: stubClient(owFixture), db: nil}
	results := p.GetMeasurement(shared.SearchRequest{})
	for _, r := range results {
		if r.Source != "OpenWeather" {
			t.Errorf("Source: got %q, want %q", r.Source, "OpenWeather")
		}
	}
}

func TestOWForecast_GetMeasurement_TypeIsForecast(t *testing.T) {
	cnf := &shared.LoggerCnf{}
	p := owForecast{cnf: cnf, l: shared.MakeNullLogger(), client: stubClient(owFixture), db: nil}
	results := p.GetMeasurement(shared.SearchRequest{})
	for _, r := range results {
		if r.Type != shared.MeasurementResult_Type_Forecast {
			t.Errorf("Type: got %q, want %q", r.Type, shared.MeasurementResult_Type_Forecast)
		}
	}
}

func TestOWForecast_GetMeasurement_AtMatchesFirstEntryDay(t *testing.T) {
	cnf := &shared.LoggerCnf{}
	p := owForecast{cnf: cnf, l: shared.MakeNullLogger(), client: stubClient(owFixture), db: nil}
	results := p.GetMeasurement(shared.SearchRequest{})

	if len(results) == 0 {
		t.Fatal("expected results, got none")
	}
	want := time.Unix(1780185600, 0).UTC().Format(time.RFC3339)
	if results[0].At != want {
		t.Errorf("At: got %q, want %q", results[0].At, want)
	}
}

func TestOWForecast_GetMeasurement_LocationPropagated(t *testing.T) {
	loc := shared.Location{Name: "Paris"}
	cnf := &shared.LoggerCnf{}
	p := owForecast{cnf: cnf, l: shared.MakeNullLogger(), client: stubClient(owFixture), db: nil}
	results := p.GetMeasurement(shared.SearchRequest{Loc: loc})
	for _, r := range results {
		if r.Loc.Name != loc.Name {
			t.Errorf("Loc.Name: got %q, want %q", r.Loc.Name, loc.Name)
		}
	}
}

func TestOWForecast_GetMeasurement_MalformedJSON(t *testing.T) {
	cnf := &shared.LoggerCnf{}
	p := owForecast{cnf: cnf, l: shared.MakeNullLogger(), client: stubClient("not json"), db: nil}
	results := p.GetMeasurement(shared.SearchRequest{})
	if len(results) != 0 {
		t.Errorf("expected 0 results for malformed JSON, got %d", len(results))
	}
}

func TestOWForecast_GetMeasurement_SavesRawResponse(t *testing.T) {
	db := tempDB(t, bucketOpenWeather)
	cnf := &shared.LoggerCnf{}
	p := owForecast{cnf: cnf, l: shared.MakeNullLogger(), client: stubClient(owFixture), db: db}
	p.GetMeasurement(shared.SearchRequest{Loc: shared.Location{Name: "Paris"}})
	if countKeys(t, db, bucketOpenWeather) != 1 {
		t.Error("expected one raw response entry saved in DB")
	}
}

func TestOWForecast_GetMeasurement_NilDB_NoPanic(t *testing.T) {
	cnf := &shared.LoggerCnf{}
	p := owForecast{cnf: cnf, l: shared.MakeNullLogger(), client: stubClient(owFixture), db: nil}
	p.GetMeasurement(shared.SearchRequest{})
}
