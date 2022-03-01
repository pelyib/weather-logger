package business

import (
	"fmt"
	"log"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

type minAndMaxLineChartBuilder struct {
	mrType string
	minDs  datasetSelector
	maxDs  datasetSelector
	r      ChartRepository
}

type datasetSelector func(c *Chart) *Dataset

func (b minAndMaxLineChartBuilder) Build(mrs []shared.MeasurementResult) {
	log.Println(fmt.Sprintf("lineChartBuilder (%s): start", b.mrType))

	for _, mr := range mrs {
		if mr.Type != b.mrType {
			log.Println(fmt.Sprintf("LineChartBuilder (%s): skipping measurement result", b.mrType))

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

		log.Println(fmt.Sprintf("LineChartBuilder (%s): %s saving", b.mrType, at.Format("2006.01")))
		b.r.Save(*chart)
	}

	log.Println(fmt.Sprintf("LineChartBuilder (%s): finished", b.mrType))
}

func MakeForecastLineChartBuilder(r *ChartRepository) ChartBuilder {
	return minAndMaxLineChartBuilder{
		mrType: shared.MeasurementResult_Type_Forecast,
		r:      *r,
		minDs:  func(c *Chart) *Dataset { return c.ForecastMinLineDataset() },
		maxDs:  func(c *Chart) *Dataset { return c.ForecastMaxLineDataset() },
	}
}

func MakeHistoricalLineChartBuilder(r *ChartRepository) ChartBuilder {
	return minAndMaxLineChartBuilder{
		mrType: shared.MeasurementResult_Type_Historical,
		r:      *r,
		minDs:  func(c *Chart) *Dataset { return c.HistoricalMinLineDataset() },
		maxDs: func(c *Chart) *Dataset {
			fmt.Println(len(c.Datasets))
			d := c.HistoricalMaxLineDataset()
			fmt.Println(len(c.Datasets))

			return d
		},
	}
}
