package in

import (
	"encoding/json"
	"fmt"

	"github.com/pelyib/weather-logger/internal/http/business"
	logger "github.com/pelyib/weather-logger/internal/logger/business"
	"github.com/pelyib/weather-logger/internal/shared/mq"
)

type executor struct {
  cb business.ChartBuilder
}

func (e executor) Execute(msg []byte) {
  fcs := []logger.Forecast{}
  err := json.Unmarshal(msg, &fcs)

  if err == nil {
    e.cb.Build(fcs)
    return
  }

  fmt.Println("UpdChartCmdExecutor | could not unmarshal message")
  fmt.Println(err)
}

func MakeUpdateChartsCommandExecutor(cb business.ChartBuilder) mq.Executor {
  return executor{cb: cb}
}
