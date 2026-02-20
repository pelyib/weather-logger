package business

import (
	"github.com/pelyib/weather-logger/internal/shared"
)

type chartBuilderFacade struct {
	builders []ChartBuilder
}

type ChartBuilder interface {
	Build(mrs []shared.MeasurementResult) error
}

func (s chartBuilderFacade) Build(mrs []shared.MeasurementResult) error {
	for _, b := range s.builders {
		if err := b.Build(mrs); err != nil {
			return err
		}
	}
	return nil
}

func MakeChartBuilder(r *ChartRepository) ChartBuilder {
	return chartBuilderFacade{
		builders: []ChartBuilder{
			MakeForecastLineChartBuilder(r, shared.MakeCliLogger(shared.App_Http, "ChartBuilder.Line.Forecast")),
			MakeHistoricalLineChartBuilder(r, shared.MakeCliLogger(shared.App_Http, "ChartBuilder.Line.Historical")),
			MakeForecastBubbleChartBuilder(r, shared.MakeCliLogger(shared.App_Http, "ChartBuilder.Bubble.Forecast")),
		},
	}
}
