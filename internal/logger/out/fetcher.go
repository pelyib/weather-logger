package out

import (
	"fmt"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

type WeatherProviderAdapter interface {
	sourceId() string
	fetch(sr shared.SearchRequest) ([]byte, error)
	MapToMeasurements(rawApiRes []byte, loc shared.Location) ([]shared.MeasurementResult, error)
}

type DbClient interface {
	saveRawApiRes(sourceId string, rawApiRes []byte)
}

type Fetcher struct {
	weatherProviderAdapter WeatherProviderAdapter
	dbClient               DbClient
	logger                 shared.Logger
}

type dbRecord struct {
	CalledAt time.Time   `json:"calledAt"`
	Source   string      `json:"source"`
	Raw      interface{} `json:"raw"`
}

func (f Fetcher) GetMeasurement(searchRequest shared.SearchRequest) []shared.MeasurementResult {
	rawApiRes, err := f.weatherProviderAdapter.fetch(searchRequest)

	if err != nil {
		f.logger.Error(fmt.Sprintf("Fetching %s failed, reason: %s", f.weatherProviderAdapter.sourceId(), err.Error()))
		return shared.MakeEmptyResults()
	}

	f.dbClient.saveRawApiRes(f.weatherProviderAdapter.sourceId(), rawApiRes)

	// TODO: handle errors here [pelyib]
	measurements, _ := f.weatherProviderAdapter.MapToMeasurements(rawApiRes, searchRequest.Loc)

	return measurements
}
