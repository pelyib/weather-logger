package business

import "github.com/pelyib/weather-logger/internal/shared"

type measurementResultProviderPool struct {
	extPrds []MeasurementResultProvider
}

type MeasurementResultProvider interface {
	GetMeasurement(searchRequest SearchRequest) []shared.MeasurementResult
}

func (pool measurementResultProviderPool) GetMeasurement(searchRequest SearchRequest) []shared.MeasurementResult {
	mrs := []shared.MeasurementResult{}

	for _, extPrd := range pool.extPrds {
		mrs = append(mrs, extPrd.GetMeasurement(searchRequest)...)
	}

	return mrs
}

func MakeMeasurementResultProviderPool(extPrds []MeasurementResultProvider) MeasurementResultProvider {
	return measurementResultProviderPool{extPrds: extPrds}
}
