package business

import (
	"log"

	"github.com/pelyib/weather-logger/internal/shared"
)

type chartBuilderFacade struct {
  builders []ChartBuilder
}

type ChartBuilder interface {
  Build(mrs []shared.MeasurementResult)
}

func (s chartBuilderFacade) Build(mrs []shared.MeasurementResult) {
  for _, b := range s.builders {
    log.Println("chartbuilder: building")
    b.Build(mrs)
  }
}

func MakeChartBuilder(r *ChartRepository) ChartBuilder {
  return chartBuilderFacade{
    builders: []ChartBuilder{
      MakeForecastLineChartBuilder(r),
      MakeHistoricalLineChartBuilder(r),
      MakeForecastBubbleChartBuilder(r),
    },
  }
}
