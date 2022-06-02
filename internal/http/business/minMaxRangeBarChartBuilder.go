package business

/*
import (
	"fmt"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

type minAndMaxRangeBarChartBuilder struct {
	mrType string
	r      ChartRepository
	l      shared.Logger
}

func (b minAndMaxRangeBarChartBuilder) Build(mrs []shared.MeasurementResult) {
	b.l.Info("barChartBuilder: start")

	for _, mr := range mrs {
		if mr.Type != b.mrType {
			b.l.Info(fmt.Sprintf("LineChartBuilder (%s): skipping measurement result", b.mrType))

			continue
		}

		at, _ := time.Parse(time.RFC3339, mr.At)

		chart := b.r.Load(ChartSearchRequest{Ym: at.Format("2006-01")})
	}

	b.l.Info("barChartBuilder: finished")
}

func MakeMinMaxRangeBarChartBuilder(
	r ChartRepository,
	l shared.Logger,
) ChartBuilder {
	return minAndMaxRangeBarChartBuilder{
		mrType: MeasurementResult_Type_Forecast,
		r:      r,
		l:      l,
	}
}
*/
