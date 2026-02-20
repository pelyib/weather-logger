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

func (e executor) Execute(msg []byte) error {
	mrs := shared.MakeEmptyResults()
	if err := json.Unmarshal(msg, &mrs); err != nil {
		return fmt.Errorf("decode measurement results: %w", err)
	}
	return e.cb.Build(mrs)
}

func MakeUpdateChartsCommandExecutor(cb business.ChartBuilder) mq.Executor {
	return executor{cb: cb}
}
