package business

import (
	"testing"

	"github.com/pelyib/weather-logger/internal/shared"
)

// stubProvider is a test double for MeasurementResultProvider.
type stubProvider struct {
	results []shared.MeasurementResult
}

func (s stubProvider) GetMeasurement(_ shared.SearchRequest) []shared.MeasurementResult {
	return s.results
}

func TestMeasurementResultProviderPool_AggregatesAll(t *testing.T) {
	loc := shared.Location{Name: "Berlin"}
	r1 := shared.MeasurementResult{Source: "A", Type: shared.MeasurementResult_Type_Forecast, Min: 1, Max: 5, Loc: loc}
	r2 := shared.MeasurementResult{Source: "B", Type: shared.MeasurementResult_Type_Historical, Min: 2, Max: 6, Loc: loc}
	r3 := shared.MeasurementResult{Source: "C", Type: shared.MeasurementResult_Type_Forecast, Min: 3, Max: 7, Loc: loc}

	pool := MakeMeasurementResultProviderPool([]MeasurementResultProvider{
		stubProvider{results: []shared.MeasurementResult{r1}},
		stubProvider{results: []shared.MeasurementResult{r2, r3}},
	})

	got := pool.GetMeasurement(shared.SearchRequest{Loc: loc})

	if len(got) != 3 {
		t.Errorf("expected 3 results, got %d", len(got))
	}
}

func TestMeasurementResultProviderPool_EmptyProviders(t *testing.T) {
	pool := MakeMeasurementResultProviderPool([]MeasurementResultProvider{})
	got := pool.GetMeasurement(shared.SearchRequest{})
	if len(got) != 0 {
		t.Errorf("expected 0 results from empty pool, got %d", len(got))
	}
}

func TestMeasurementResultProviderPool_PassesSearchRequest(t *testing.T) {
	want := shared.SearchRequest{Loc: shared.Location{Name: "Munich"}}
	var got shared.SearchRequest

	capturing := &capturingProvider{capture: &got}
	pool := MakeMeasurementResultProviderPool([]MeasurementResultProvider{capturing})
	pool.GetMeasurement(want)

	if got.Loc.Name != want.Loc.Name {
		t.Errorf("SearchRequest not forwarded: got %v, want %v", got, want)
	}
}

type capturingProvider struct {
	capture *shared.SearchRequest
}

func (c *capturingProvider) GetMeasurement(sr shared.SearchRequest) []shared.MeasurementResult {
	*c.capture = sr
	return nil
}
