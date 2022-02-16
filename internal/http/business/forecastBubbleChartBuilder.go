package business

import (
	"fmt"
	"log"
	"time"

	"github.com/pelyib/weather-logger/internal/logger/business"
)

const bubbleR int8 = 2

type forecastBubbleChartBuilder struct {
  repository ChartRepository
}

func (b forecastBubbleChartBuilder) Build(fcs []business.Forecast) {
  for _, fc := range fcs {
    at, _ := time.Parse(time.RFC3339, fc.At)

    chart := b.repository.Load(ChartSearchRequest{Ym: at.Format("2006-01")})

    dataset, err := chart.ForecastBubbleDataset()
    if err != nil {
      log.Println("ForecastBubbleChartBuilder: dataset is missing")
      return
    }

    minKey := fmt.Sprintf("%s_%f", at.Format("02"), fc.Min)
    maxKey := fmt.Sprintf("%s_%f", at.Format("02"), fc.Max)

    if v, ok := dataset.Data[minKey]; ok {
      dataset.Data[minKey] = Item{X: v.X, Y: v.Y, R: v.R + bubbleR}
    } else {
      dataset.Data[minKey] = Item{X: at.UnixMilli(), Y: fc.Min, R: bubbleR}
    }

    if v, ok := dataset.Data[maxKey]; ok {
      dataset.Data[maxKey] = Item{X: v.X, Y: v.Y, R: bubbleR}
    } else {
      dataset.Data[maxKey] = Item{X: at.UnixMilli(), Y: fc.Max, R: bubbleR}
    }

    b.repository.Save(chart)
  }
}

func MakeForecastBubbleChartBuilder(r *ChartRepository) ChartBuilder {
  return forecastBubbleChartBuilder{repository: *r}
}
