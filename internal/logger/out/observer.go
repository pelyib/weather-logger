package out

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pelyib/weather-logger/internal/logger/business"
	amqp "github.com/rabbitmq/amqp091-go"
)

type cliObs struct {}

type httpObs struct {
  channel *amqp.Channel
}

func (cli cliObs) Notify(forecasts []business.Forecast) {
  fmt.Println(fmt.Sprintf("CLI observer | %d forecasts fetched", len(forecasts)))
}

func (http httpObs) Notify(forecasts []business.Forecast) {
  body, _ := json.Marshal(forecasts)

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

func MakeCliObserver() business.Observer {
  return cliObs{}
}

func MakeHttpObsercer(c *amqp.Channel) business.Observer {
  return httpObs{c}
}
