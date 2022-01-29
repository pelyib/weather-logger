package business

import (
	"log"
	"time"

	"github.com/pelyib/weather-logger/internal"
)
type minAndMaxLineChartBuilder struct {
  repository ChartRepository
}

func (b minAndMaxLineChartBuilder) Build(fcs []internal.Forecast) {
  log.Println("lineChartBuilder: start building")

  for _, fc := range fcs {
    at, _ := time.Parse(time.RFC3339, fc.At)

    chart := b.repository.Load(ChartSearchRequest{Ym: at.Format("2006-01")})

    dataset, err := chart.MinLineDataset()
    if err != nil {
      log.Println("lineChartBuilder: Line dataset is missing")
      return
    }

    if v, ok := dataset.Data[at.Format("02")]; ok {
      if v.Y < fc.Min {
        dataset.Data[at.Format("02")] = Item{X: v.X, Y: fc.Min}
      }
    } else {
      at, _ := time.Parse(time.RFC3339, fc.At)
      dataset.Data[at.Format("02")] = Item{X: at.UnixMilli(), Y: fc.Min}
    }
    log.Println("lineChartBuilder: saving")
    b.repository.Save(chart)
  }
}

func MakeLineChartBuilder(r ChartRepository) chartBuilder {
  return minAndMaxLineChartBuilder{
    repository: r,
  }
}
