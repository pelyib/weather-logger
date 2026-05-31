package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"regexp"
	"syscall"
	"time"

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
	http.NotFound(w, r)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cnf, err := shared.CreateHttpConf(shared.MakeCliLogger(shared.App_Http, "Config"))
	if err != nil {
		log.Fatalln(err)
	}

	db, err := internal.MakeDb(&cnf.Database, shared.MakeCliLogger(shared.App_Http, "DB"))
	if err != nil {
		log.Fatalln(err)
	}

	cr := out.MakeChartRepository(db, shared.MakeCliLogger(shared.App_Http, "ChartRepository"))

	go func() {
		if err := consume(ctx, cnf, &cr); err != nil {
			log.Printf("MQ consumer error: %v", err)
		}
	}()

	serve(ctx, cnf, &cr)
}

func serve(ctx context.Context, cnf *shared.HttpCnf, cr *business.ChartRepository) {
	h := &regexpHandler{}

	h.HandleFunc(regexp.MustCompile("^/health$"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	hh, err := in.MakeHistoryHandler(cnf, cr)
	if err != nil {
		log.Fatalln(err)
	}
	h.HandleFunc(regexp.MustCompile("/[a-z]{2}/[a-z]{1,}/[0-9]{4}/[0-9]{2}"), func(rw http.ResponseWriter, r *http.Request) {
		hh.Handle(rw, r)
	})

	ih, err := in.MakeIndexHandler(cnf, cr)
	if err != nil {
		log.Fatalln(err)
	}
	h.HandleFunc(regexp.MustCompile("/"), func(rw http.ResponseWriter, r *http.Request) {
		ih.Handle(rw, r)
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cnf.Port),
		Handler: h,
	}

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("HTTP server error: %v", err)
	}
}

func consume(ctx context.Context, cnf *shared.HttpCnf, cr *business.ChartRepository) error {
	c, err := mq.MakeChannel(cnf.Mq, shared.MakeCliLogger(shared.App_Http, "MQ"))
	if err != nil {
		return err
	}

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

	return cons.Consume(ctx)
}
