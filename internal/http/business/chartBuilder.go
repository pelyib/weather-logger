package business

import (
	"log"

	"github.com/pelyib/weather-logger/internal/logger/business"
)

type chartBuilderFacade struct {
  builders []ChartBuilder
}

type ChartBuilder interface {
  Build(fcs []business.Forecast)
}

func (s chartBuilderFacade) Build(fcs []business.Forecast) {
  for _, b := range s.builders {
    log.Println("chartbuilder: building")
    b.Build(fcs)
  }
}

func MakeChartBuilder(r *ChartRepository) ChartBuilder {
  return chartBuilderFacade{
    builders: []ChartBuilder{
      MakeLineChartBuilder(r),
      MakeForecastBubbleChartBuilder(r),
    },
  }
}
