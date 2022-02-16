package business

import (
	"fmt"
	"log"
	"time"

	"github.com/pelyib/weather-logger/internal/logger/business"
)
type minAndMaxLineChartBuilder struct {
  repository ChartRepository
}

func (b minAndMaxLineChartBuilder) Build(fcs []business.Forecast) {
  log.Println("lineChartBuilder: start building")

  for _, fc := range fcs {
    at, _ := time.Parse(time.RFC3339, fc.At)

    chart := b.repository.Load(ChartSearchRequest{Ym: at.Format("2006-01")})

    min, err := chart.MinLineDataset()
    if err != nil {
      log.Println("lineChartBuilder | MinLineDataset is missing")
      return
    }

    max, err := chart.MaxLineDataset()
    if err != nil {
      log.Println("lineChartBuilder | MaxLineDataset is missing")
    }

    if v, ok := min.Data[at.Format("02")]; ok {
      if v.Y > fc.Min {
        min.Data[at.Format("02")] = Item{X: v.X, Y: fc.Min}
      }
    } else {
      at, _ := time.Parse(time.RFC3339, fc.At)
      min.Data[at.Format("02")] = Item{X: at.UnixMilli(), Y: fc.Min}
    }

    if v, ok := max.Data[at.Format("02")]; ok {
      if v.Y < fc.Max {
        max.Data[at.Format("02")] = Item{X: v.X, Y: fc.Max}
      }
    } else {
      at, _ := time.Parse(time.RFC3339, fc.At)
      max.Data[at.Format("02")] = Item{X: at.UnixMilli(), Y: fc.Max}
    }

    log.Println(fmt.Sprintf("lineChartBuilder | %s saving", at.Format("2006.01")))
    b.repository.Save(chart)
  }
}

func MakeLineChartBuilder(r *ChartRepository) ChartBuilder {
  return minAndMaxLineChartBuilder{
    repository: *r,
  }
}
