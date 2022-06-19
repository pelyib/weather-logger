package in

import (
	"encoding/json"
	"fmt"

	"github.com/pelyib/weather-logger/internal/logger/business"
	"github.com/pelyib/weather-logger/internal/shared"
	"github.com/pelyib/weather-logger/internal/shared/mq"
)

type fetchCommandExecutor struct {
	mrp business.MeasurementResultProvider
	obs []business.Observer
}

func (executor fetchCommandExecutor) Execute(msg []byte) {
	msgDecoded := mq.MsgBody{}
	err := json.Unmarshal(msg, &msgDecoded)

	if err != nil {
		fmt.Println("banan")
		fmt.Println(err)
		return
	}

	measurementResults := executor.mrp.GetMeasurement(shared.SearchRequest{Loc: msgDecoded.Loc})

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
