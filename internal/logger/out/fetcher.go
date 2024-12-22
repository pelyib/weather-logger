package out

import (
	"fmt"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

type WeatherProviderAdapter interface {
	SourceId() string
	Fetch(sr shared.SearchRequest) ([]byte, error)
	MapToMeasurements(rawApiRes []byte, loc shared.Location) []shared.MeasurementResult
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
	rawApiRes, err := f.weatherProviderAdapter.Fetch(searchRequest)

	if err != nil {
		f.logger.Error(fmt.Sprintf("Fetching %s failed, reason: %s", f.weatherProviderAdapter.SourceId(), err.Error()))
		return shared.MakeEmptyResults()
	}

	f.dbClient.saveRawApiRes(f.weatherProviderAdapter.SourceId(), rawApiRes)

	return f.weatherProviderAdapter.MapToMeasurements(rawApiRes, searchRequest.Loc)
}
