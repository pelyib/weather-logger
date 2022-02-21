package out

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pelyib/weather-logger/internal/logger/business"
	"github.com/pelyib/weather-logger/internal/shared"
	amqp "github.com/rabbitmq/amqp091-go"
)

type cliObs struct {
  verbose bool
}

type httpObs struct {
  channel *amqp.Channel
}

func (cli cliObs) Notify(mrs []shared.MeasurementResult) {
  fmt.Println(fmt.Sprintf("CLI observer | %d measurements results fethced", len(mrs)))

  if cli.verbose {
    json, _ := json.Marshal(mrs)
    fmt.Println(string(json))
  }
}

func (http httpObs) Notify(mrs []shared.MeasurementResult) {
  body, _ := json.Marshal(mrs)

  msg := amqp.Publishing{
    DeliveryMode: amqp.Persistent,
    Timestamp:    time.Now(),
    ContentType:  "application/json",
    Body:         body,
  }

  err := http.channel.Publish("http", "update:charts", false, false, msg)

  if err != nil {
    fmt.Println("HTTP Observer | Could not send MQ message")
  } else {
    fmt.Println("HTTP Observer | published")
  }
}

func MakeCliObserver(verbose bool) business.Observer {
  return cliObs{verbose: verbose}
}

func MakeHttpObserver(c *amqp.Channel) business.Observer {
  return httpObs{c}
}
