package main

import (
	"fmt"
	"log"

	"github.com/pelyib/weather-logger/internal"
	"github.com/pelyib/weather-logger/internal/http/business"
	"github.com/pelyib/weather-logger/internal/http/out"
)

var cnf internal.Cnf

func main() {
  cnf, err := internal.CreateLoggerConf()

  if err != nil {
    log.Fatalln(err)
  }
fmt.Println("loading DB")
  db := internal.MakeDb(cnf)
fmt.Println("DB loaded")
  fss := MakeFss(cnf, &db) // TODO inject DB
fmt.Println("FC service loaded")
  fcs := fss.get(internal.SearchRequest{})

  log.Println("main: Forecasts fetched")
  cb := business.MakeChartBuilder(out.MakeChartRepository(&db))
  log.Println("main: chartbuilder built")
  cb.Build(fcs)
  log.Println("main: chartbuilder done")
  add(fcs)
}
