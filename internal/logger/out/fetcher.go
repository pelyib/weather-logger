package out

import (
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

type Adapter interface {
	SourceId() string
	Fetch(sr shared.SearchRequest) []byte
	MapToMeasurements(rawApiRes []byte) []shared.MeasurementResult
}

type Fetcher struct {
	weatherProviderAdapter Adapter
	dbClient               client
}

type dbRecord struct {
	CalledAt time.Time   `json:"calledAt"`
	Source   string      `json:"source"`
	Raw      interface{} `json:"raw"`
}

func (f Fetcher) GetMeasurement(searchRequest shared.SearchRequest) []shared.MeasurementResult {
	rawApiRes := f.weatherProviderAdapter.Fetch(searchRequest)

	f.dbClient.saveRawApiRes(f.weatherProviderAdapter.SourceId(), rawApiRes)

	return f.weatherProviderAdapter.MapToMeasurements(rawApiRes)
}
