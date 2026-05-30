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

func (executor fetchCommandExecutor) Execute(msg []byte) error {
	msgDecoded := mq.MsgBody{}
	if err := json.Unmarshal(msg, &msgDecoded); err != nil {
		return fmt.Errorf("decode message body: %w", err)
	}

	measurementResults := executor.mrp.GetMeasurement(shared.SearchRequest{Loc: msgDecoded.Loc})

	for _, observer := range executor.obs {
		observer.Notify(measurementResults)
	}

	return nil
}

func MakeFetchCommandExecutor(mrp business.MeasurementResultProvider, obs []business.Observer) mq.Executor {
	return fetchCommandExecutor{mrp, obs}
}
