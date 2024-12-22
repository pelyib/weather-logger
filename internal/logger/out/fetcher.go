package out

import (
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

type WeatherProviderAdapter interface {
	SourceId() string
	Fetch(sr shared.SearchRequest) []byte
	MapToMeasurements(rawApiRes []byte) []shared.MeasurementResult
}

type DbClient interface {
	saveRawApiRes(sourceId string, rawApiRes []byte)
}

type Fetcher struct {
	weatherProviderAdapter WeatherProviderAdapter
	dbClient               DbClient
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
