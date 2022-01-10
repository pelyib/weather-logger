package out

import "github.com/pelyib/weather-logger/internal"

type ExternalProvider interface {
  Get(s internal.SearchRequest) []internal.Forecast
}
