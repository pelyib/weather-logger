package in

import (
	"github.com/pelyib/weather-logger/internal/logger/business"
	"github.com/pelyib/weather-logger/internal/shared/mq"
)

type fetchCommandExecutor struct {
	mrp business.MeasurementResultProvider
	obs []business.Observer
}

func (executor fetchCommandExecutor) Execute(msg []byte) {
	measurementResults := executor.mrp.GetMeasurement(business.SearchRequest{}) // TODO: Make SearchRequest from msq [botond.pelyi]

	for _, observer := range executor.obs {
		observer.Notify(measurementResults)
	}
}

func MakeFetchCommandExecutor(mrp business.MeasurementResultProvider, obs []business.Observer) mq.Executor {
	return fetchCommandExecutor{
		mrp,
		obs,
	}
}
