package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
	"github.com/pelyib/weather-logger/internal/shared/mq"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	l := shared.MakeCliLogger(shared.App_Commander, "Commander")
	execute := flag.String("execute", "help", "Command to be executed")
	flag.Parse()

	l.Info(*execute)

	switch *execute {
	case "forecasts", "historical":
		cnf, err := shared.CreateLoggerConf(shared.MakeCliLogger(shared.App_Commander, "Config"))
		if err != nil {
			l.Error(fmt.Sprintf("Could not create configuration, reason: %s", err.Error()))
			os.Exit(2)
		}

		executeFetchCommands(*execute, cnf.Mq, shared.MakeCliLogger(shared.App_Commander, "FetchCommands"))
	case "help":
		l.Info("Unknow command")
	}
}

func executeFetchCommands(cmd string, cnf shared.Mq, l shared.Logger) {
	c := mq.MakeChannel(cnf, shared.MakeCliLogger(shared.App_Commander, "MQ"))

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
	}
	routingKey := fmt.Sprintf("fetch:%s", cmd)
	err := c.Publish("logger", routingKey, false, false, msg)

	if err != nil {
		l.Warning("Could not send message")
	} else {
		l.Info(fmt.Sprintf("Message (%s) published", routingKey))
	}

	c.Close()
}
