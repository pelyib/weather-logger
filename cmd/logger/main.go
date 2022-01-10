package main

import (
	"fmt"
	"log"

	"github.com/pelyib/weather-logger/internal"
)

var cnf internal.Cnf

func main() {
  cnf, err := internal.CreateLoggerConf()

  if err != nil {
    log.Fatalln(err)
  }

  fss := MakeFss(cnf)

  fs := fss.get(internal.SearchRequest{})

  fmt.Println("main: processed")

  add(fs)
}
