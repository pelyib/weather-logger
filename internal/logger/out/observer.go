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
	l       shared.Logger
}

type httpObs struct {
	channel *amqp.Channel
	l       shared.Logger
}

func (cli cliObs) Notify(mrs []shared.MeasurementResult) {
	cli.l.Info(fmt.Sprintf("%d measurements results fethced", len(mrs)))

	if cli.verbose {
		json, _ := json.Marshal(mrs)
		cli.l.Info(string(json))
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
		http.l.Warning("Could not send message")
	} else {
		http.l.Info("Message published")
	}
}

func MakeCliObserver(verbose bool, l shared.Logger) business.Observer {
	return cliObs{verbose: verbose, l: l}
}

func MakeHttpObserver(c *amqp.Channel, l shared.Logger) business.Observer {
	return httpObs{channel: c, l: l}
}
