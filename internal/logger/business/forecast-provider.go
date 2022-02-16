package business

type forecastProviderPool struct {
  extPrds []ForecastProvider
}

type ForecastProvider interface {
  Get(searchRequest SearchRequest) []Forecast
}

func (s forecastProviderPool) Get(searchRequest SearchRequest) []Forecast {
  fcs := []Forecast{}

  for _, extPrd := range s.extPrds {
    fcs = append(fcs, extPrd.Get(searchRequest)...)
  }

  return fcs
}

func MakeForecastProvderPool(extPrds []ForecastProvider) ForecastProvider {
  return forecastProviderPool{extPrds: extPrds}
}

