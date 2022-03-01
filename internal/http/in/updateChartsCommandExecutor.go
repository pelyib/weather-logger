package in

import (
	"encoding/json"
	"fmt"

	"github.com/pelyib/weather-logger/internal/http/business"
	"github.com/pelyib/weather-logger/internal/shared"
	"github.com/pelyib/weather-logger/internal/shared/mq"
)

type executor struct {
	cb business.ChartBuilder
}

func (e executor) Execute(msg []byte) {
	mrs := []shared.MeasurementResult{}
	err := json.Unmarshal(msg, &mrs)

	if err == nil {
		e.cb.Build(mrs)
		return
	}

	fmt.Println("UpdChartCmdExecutor | could not unmarshal message")
	fmt.Println(err)
}

func MakeUpdateChartsCommandExecutor(cb business.ChartBuilder) mq.Executor {
	return executor{cb: cb}
}
