package in

import (
	"github.com/pelyib/weather-logger/internal/logger/business"
	"github.com/pelyib/weather-logger/internal/shared/mq"
)

type fetchForecastCommandExecutor struct {
  forecastProvider business.ForecastProvider
  observers []business.Observer
}

func (executor fetchForecastCommandExecutor) Execute(msg []byte) {
  fcs := executor.forecastProvider.Get(business.SearchRequest{}) // TODO: Make SearchRequest from msq [botond.pelyi]

  for _, observer := range executor.observers {
    observer.Notify(fcs)
  }
}

func MakeFetchForecastCommandExecutor(fcPrd business.ForecastProvider, obs []business.Observer) mq.Executor {
  return fetchForecastCommandExecutor{
    fcPrd,
    obs,
  }
}
