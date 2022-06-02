package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/pelyib/weather-logger/internal"
	"github.com/pelyib/weather-logger/internal/http/business"
	"github.com/pelyib/weather-logger/internal/http/in"
	"github.com/pelyib/weather-logger/internal/http/out"
	"github.com/pelyib/weather-logger/internal/shared"
	"github.com/pelyib/weather-logger/internal/shared/mq"
)

// https://stackoverflow.com/questions/6564558/wildcards-in-the-pattern-for-http-handlefunc
type route struct {
	pattern *regexp.Regexp
	handler http.Handler
}

type regexpHandler struct {
	routes []*route
}

func (h *regexpHandler) Handler(pattern *regexp.Regexp, handler http.Handler) {
	h.routes = append(h.routes, &route{pattern, handler})
}

func (h *regexpHandler) HandleFunc(pattern *regexp.Regexp, handler func(http.ResponseWriter, *http.Request)) {
	h.routes = append(h.routes, &route{pattern, http.HandlerFunc(handler)})
}

func (h *regexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	// no pattern matched; send 404 response
	http.NotFound(w, r)
}

func main() {
	cnf, err := shared.CreateHttpConf(shared.MakeCliLogger(shared.App_Http, "Config"))

	if err != nil {
		log.Fatalln(err)
	}

	db := internal.MakeDb(&cnf.Database, shared.MakeCliLogger(shared.App_Http, "DB"))
	cr := out.MakeChartRepository(
		&db,
		shared.MakeCliLogger(shared.App_Http, "ChartRepository"),
	)

	go func() {
		consume(cnf, &cr)
	}()

	serve(cnf, &cr)
}

func serve(cnf *shared.HttpCnf, cr *business.ChartRepository) {
	h := &regexpHandler{}

	hh := in.MakeHistoryHandler(cnf, cr)
	hp, _ := regexp.Compile("/de/trier/[0-9]{4}/[0-9]{2}")
	h.HandleFunc(hp, func(rw http.ResponseWriter, r *http.Request) {
		hh.Handle(rw, r)
	})

	ih := in.MakeIndexHandler(cnf, cr)
	ip, _ := regexp.Compile("/")
	h.HandleFunc(ip, func(rw http.ResponseWriter, r *http.Request) {
		ih.Handle(rw, r)
	})

	http.ListenAndServe(fmt.Sprintf(":%d", cnf.Port), h)
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
		L: shared.MakeCliLogger(shared.App_Http, "MQ.Consumer"),
	}

	cons.Consume()
}
