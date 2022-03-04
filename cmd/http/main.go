package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pelyib/weather-logger/internal"
	"github.com/pelyib/weather-logger/internal/http/business"
	"github.com/pelyib/weather-logger/internal/http/in"
	"github.com/pelyib/weather-logger/internal/http/out"
	"github.com/pelyib/weather-logger/internal/shared"
	"github.com/pelyib/weather-logger/internal/shared/mq"
)

func main() {
	cnf, err := shared.CreateHttpConf(shared.MakeCliLogger("http", "Config"))

	if err != nil {
		log.Fatalln(err)
	}

	db := internal.MakeDb(&cnf.Database, shared.MakeCliLogger("http", "DB"))
	cr := out.MakeChartRepository(&db)

	go func() {
		consume(cnf, &cr)
	}()

	ih := in.MakeIndexHandler(&cr)

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		ih.Index(rw, r)
	})

	http.ListenAndServe(fmt.Sprintf(":%d", cnf.Port), nil)
}

func consume(cnf *shared.HttpCnf, cr *business.ChartRepository) {
	c := mq.MakeChannel(cnf.Mq, shared.MakeCliLogger(shared.App_Http, "MQ"))

	cons := mq.Consumer{
		Exchange: "http",
		Handlers: map[string]mq.Executor{
			"update:charts": in.MakeUpdateChartsCommandExecutor(
				business.MakeChartBuilder(cr),
			),
		},
		C: c,
	}

	cons.Consume()
}
