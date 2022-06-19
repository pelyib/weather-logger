package business

import (
	"fmt"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

const bubbleR int8 = 2

type forecastBubbleChartBuilder struct {
	repository ChartRepository
	l          shared.Logger
}

func (b forecastBubbleChartBuilder) Build(mrs []shared.MeasurementResult) {
	b.l.Info("(forecast): start building")
	for _, mr := range mrs {
		if mr.Type != shared.MeasurementResult_Type_Forecast {
			continue
		}

		at, _ := time.Parse(time.RFC3339, mr.At)

		chart := b.repository.Load(ChartSearchRequest{Ym: at.Format("2006-01"), Loc: mr.Loc})

		dataset := chart.ForecastBubbleDataset()
		minKey := fmt.Sprintf("%s_%f", at.Format("02"), mr.Min)
		maxKey := fmt.Sprintf("%s_%f", at.Format("02"), mr.Max)

		if v, ok := dataset.Data[minKey]; ok {
			dataset.Push(minKey, Item{X: v.X, Y: v.Y, R: v.R + bubbleR})
		} else {
			dataset.Push(minKey, Item{X: at.UnixMilli(), Y: mr.Min, R: bubbleR})
		}

		if v, ok := dataset.Data[maxKey]; ok {
			dataset.Push(maxKey, Item{X: v.X, Y: v.Y, R: v.R + bubbleR})
		} else {
			dataset.Push(maxKey, Item{X: at.UnixMilli(), Y: mr.Max, R: bubbleR})
		}

		b.repository.Save(*chart)
	}

	b.l.Info("(forecast): building finished")
}

func MakeForecastBubbleChartBuilder(r *ChartRepository, l shared.Logger) ChartBuilder {
	return forecastBubbleChartBuilder{repository: *r, l: l}
}
