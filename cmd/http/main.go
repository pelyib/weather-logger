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

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
  cnf, err := shared.CreateHttpConf()

  if err != nil {
    log.Fatalln(err)
  }

  db := internal.MakeDb(&cnf.Database)
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
  conn, err := amqp.Dial(
    fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
      cnf.Mq.User,
      cnf.Mq.Password,
      cnf.Mq.Domain,
      cnf.Mq.Port,
      cnf.Mq.Vhost,
    ),
  )

  if err != nil {
    log.Fatalf("connection.open: %s", err)
  }

  defer conn.Close()

  c, err := conn.Channel()
  if err != nil {
    log.Fatalf("channel.open: %s", err)
  }

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
