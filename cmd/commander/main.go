package main

import (
	"encoding/json"
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

	switch *execute {
	case "forecasts", "historical":
		cnf, err := shared.CreateLoggerConf(shared.MakeCliLogger(shared.App_Commander, "Config"))
		if err != nil {
			l.Error(fmt.Sprintf("Could not create configuration, reason: %s", err.Error()))
			os.Exit(2)
		}

		executeFetchCommands(*execute, cnf, shared.MakeCliLogger(shared.App_Commander, "FetchCommands"))
	case "help":
		l.Info("Unknown command (given:" + *execute + ")\n\nPossible flags\n  - forecasts  : will trigger the logger to fetch forecasts from the providers\n  - historical : will trigger the logger to fetch historical data from providers")
	}
}

func executeFetchCommands(cmd string, cnf *shared.LoggerCnf, l shared.Logger) {
	c := mq.MakeChannel(cnf.Mq, shared.MakeCliLogger(shared.App_Commander, "MQ"))

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
	}
	routingKey := fmt.Sprintf("fetch:%s", cmd)

	for _, loc := range cnf.Locations {
		msgJson, err := json.Marshal(mq.MsgBody{Loc: loc})

		if err == nil {
			msg.Body = msgJson
			err = c.Publish("logger", routingKey, false, false, msg)

			if err == nil {
				l.Info(fmt.Sprintf("Message published | cmd: %s | loc: %s-%s", routingKey, loc.Country.Alpha2Code, loc.Name))
			} else {
				l.Warning("Could not send message")
			}
		} else {
			l.Warning("Could not create message body")
		}
	}

	c.Close()
}
