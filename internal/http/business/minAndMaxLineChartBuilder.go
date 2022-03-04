package business

import (
	"fmt"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

type minAndMaxLineChartBuilder struct {
	mrType string
	minDs  datasetSelector
	maxDs  datasetSelector
	r      ChartRepository
	l      shared.Logger
}

type datasetSelector func(c *Chart) *Dataset

func (b minAndMaxLineChartBuilder) Build(mrs []shared.MeasurementResult) {
	b.l.Info(fmt.Sprintf("lineChartBuilder (%s): start", b.mrType))

	for _, mr := range mrs {
		if mr.Type != b.mrType {
			b.l.Info(fmt.Sprintf("LineChartBuilder (%s): skipping measurement result", b.mrType))

			continue
		}

		at, _ := time.Parse(time.RFC3339, mr.At)

		chart := b.r.Load(ChartSearchRequest{Ym: at.Format("2006-01")})

		min := b.minDs(chart)
		max := b.maxDs(chart)

		if v, ok := min.Data[at.Format("02")]; ok {
			if v.Y > mr.Min {
				min.Push(at.Format("02"), Item{X: v.X, Y: mr.Min})
			}
		} else {
			at, _ := time.Parse(time.RFC3339, mr.At)
			min.Push(at.Format("02"), Item{X: at.UnixMilli(), Y: mr.Min})
		}

		if v, ok := max.Data[at.Format("02")]; ok {
			if v.Y < mr.Max {
				max.Push(at.Format("02"), Item{X: v.X, Y: mr.Max})
			}
		} else {
			at, _ := time.Parse(time.RFC3339, mr.At)
			max.Push(at.Format("02"), Item{X: at.UnixMilli(), Y: mr.Max})
		}

		b.l.Info(fmt.Sprintf("(%s): %s saving", b.mrType, at.Format("2006.01")))
		b.r.Save(*chart)
	}

	b.l.Info(fmt.Sprintf("(%s): finished", b.mrType))
}

func MakeForecastLineChartBuilder(r *ChartRepository, l shared.Logger) ChartBuilder {
	return minAndMaxLineChartBuilder{
		mrType: shared.MeasurementResult_Type_Forecast,
		r:      *r,
		minDs:  func(c *Chart) *Dataset { return c.ForecastMinLineDataset() },
		maxDs:  func(c *Chart) *Dataset { return c.ForecastMaxLineDataset() },
		l:      l,
	}
}

func MakeHistoricalLineChartBuilder(r *ChartRepository, l shared.Logger) ChartBuilder {
	return minAndMaxLineChartBuilder{
		mrType: shared.MeasurementResult_Type_Historical,
		r:      *r,
		minDs:  func(c *Chart) *Dataset { return c.HistoricalMinLineDataset() },
		maxDs:  func(c *Chart) *Dataset { return c.HistoricalMaxLineDataset() },
		l:      l,
	}
}
