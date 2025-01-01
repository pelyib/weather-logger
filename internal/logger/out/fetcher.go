package out

import (
	"fmt"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

type WeatherProviderAdapter interface {
	sourceId() string
	fetch(sr shared.SearchRequest) ([]byte, error)
	mapToMeasurements(rawApiRes []byte, loc shared.Location) ([]shared.MeasurementResult, error)
}

type DbClient interface {
	saveRawApiRes(sourceId string, rawApiRes []byte) error
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

	err = f.dbClient.saveRawApiRes(f.weatherProviderAdapter.sourceId(), rawApiRes)

	if err != nil {
		f.logger.Error(fmt.Sprintf("Saving raw API response failed, reason: %s", err.Error()))
	}

	measurements, err := f.weatherProviderAdapter.mapToMeasurements(rawApiRes, searchRequest.Loc)

	if err != nil {
		f.logger.Error(fmt.Sprintf("Mapping raw API response to measurements failed, reason: %s", err.Error()))

		return shared.MakeEmptyResults()
	}

	return measurements
}
