package main

import (
	"github.com/pelyib/weather-logger/internal"
	"github.com/pelyib/weather-logger/internal/logger/out"
)

type ForecastProviderService struct {
  cnf *internal.LoggerCnf
  extPrds []out.ExternalProvider
}

type ForecastProvider interface {
  get(searchRequest internal.SearchRequest) []internal.Forecast
}

func (s ForecastProviderService) get(searchRequest internal.SearchRequest) []internal.Forecast {
  fcs := []internal.Forecast{}

  for _, extPrd := range s.extPrds {
    fcs = append(fcs, extPrd.Get(searchRequest)...)
  }

  return fcs
}

func MakeFss(cnf *internal.LoggerCnf) ForecastProvider {
  providers := []out.ExternalProvider{out.MakeAW(cnf), out.MakeOW(cnf)}

  return ForecastProviderService{cnf: cnf, extPrds: providers}
}

