package business

import (
	"fmt"
	"log"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

const bubbleR int8 = 2

type forecastBubbleChartBuilder struct {
	repository ChartRepository
}

func (b forecastBubbleChartBuilder) Build(mrs []shared.MeasurementResult) {

	log.Println(fmt.Sprintf("Bubblebuilder: star"))
	for _, mr := range mrs {
		if mr.Type != shared.MeasurementResult_Type_Forecast {
			continue
		}

		at, _ := time.Parse(time.RFC3339, mr.At)

		chart := b.repository.Load(ChartSearchRequest{Ym: at.Format("2006-01")})

		dataset := chart.ForecastBubbleDataset()
		minKey := fmt.Sprintf("%s_%f", at.Format("02"), mr.Min)
		maxKey := fmt.Sprintf("%s_%f", at.Format("02"), mr.Max)

		if v, ok := dataset.Data[minKey]; ok {
			dataset.Push(minKey, Item{X: v.X, Y: v.Y, R: v.R + bubbleR})
		} else {
			dataset.Push(minKey, Item{X: at.UnixMilli(), Y: mr.Min, R: bubbleR})
		}

		if v, ok := dataset.Data[maxKey]; ok {
			dataset.Push(maxKey, Item{X: v.X, Y: v.Y, R: bubbleR})
		} else {
			dataset.Push(maxKey, Item{X: at.UnixMilli(), Y: mr.Max, R: bubbleR})
		}

		b.repository.Save(*chart)
	}

	log.Println("Bubblebuilder: finished")
}

func MakeForecastBubbleChartBuilder(r *ChartRepository) ChartBuilder {
	return forecastBubbleChartBuilder{repository: *r}
}
