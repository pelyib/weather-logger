package business

import (
	"log"

	"github.com/pelyib/weather-logger/internal"
)

type chartBuilderFacade struct {
  builders []chartBuilder
}

type chartBuilder interface {
  Build(fcs []internal.Forecast)
}

func (s chartBuilderFacade) Build(fcs []internal.Forecast) {
  for _, b := range s.builders {
    log.Println("chartbuilder: building")
    b.Build(fcs)
  }
}

func MakeChartBuilder(r ChartRepository) chartBuilder {
  return chartBuilderFacade{
    builders: []chartBuilder{
      MakeLineChartBuilder(r),
      MakeForecastBubbleChartBuilder(r),
    },
  }
}
