package main

import (
  bolt "go.etcd.io/bbolt"

  "github.com/pelyib/weather-logger/internal"
  "github.com/pelyib/weather-logger/internal/logger/out"
)

type ForecastProviderService struct {
  cnf *internal.LoggerCnf
  extPrds []out.ExternalProvider
  db bolt.DB
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

func MakeFss(cnf *internal.LoggerCnf, db *bolt.DB) ForecastProvider {
  providers := []out.ExternalProvider{out.MakeAW(cnf, db), out.MakeOW(cnf)}
  return ForecastProviderService{cnf: cnf, extPrds: providers, db: *db}
}

