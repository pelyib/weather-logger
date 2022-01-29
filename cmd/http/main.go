package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pelyib/weather-logger/internal"
	"github.com/pelyib/weather-logger/internal/http/out"
)

func main() {
  cnf, err := internal.CreateHttpConf()

  if err != nil {
    log.Fatalln(err)
  }

  // Temporary, Database should get its own config struct [botond.pelyi]
  lcnf, err := internal.CreateLoggerConf()

  if err != nil {
    log.Fatalln(err)
  }

  db := internal.MakeDb(lcnf) // TODO: it is read-write but we need ONLY read [botond.pelyi]

  ih := MakeIndexHandler(out.MakeChartRepository(&db))

  http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
    ih.index(rw, r)
  })

  http.ListenAndServe(fmt.Sprintf(":%d", cnf.Port), nil)
}
