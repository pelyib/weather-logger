package main

import (
	"fmt"
	"log"

	"github.com/pelyib/weather-logger/internal"
	"github.com/pelyib/weather-logger/internal/logger/business"
	"github.com/pelyib/weather-logger/internal/logger/in"
	"github.com/pelyib/weather-logger/internal/logger/out"
	"github.com/pelyib/weather-logger/internal/shared"
	"github.com/pelyib/weather-logger/internal/shared/mq"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	cnf, err := shared.CreateLoggerConf()

	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("loading DB")
	db := internal.MakeDb(&cnf.Database)
	fmt.Println("DB loaded")

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
			business.COMMAND_FETCH_FORECASTS: in.MakeFetchCommandExecutor(
				business.MakeMeasurementResultProviderPool([]business.MeasurementResultProvider{
					out.MakeAccuWeatherForecastProvider(cnf, &db),
					out.MakeOW(cnf),
				}),
				[]business.Observer{out.MakeCliObserver(false), out.MakeHttpObserver(c)},
			),
			business.COMMAND_FETCH_HISTORICAL: in.MakeFetchCommandExecutor(
				business.MakeMeasurementResultProviderPool(
					[]business.MeasurementResultProvider{
						out.MakeAccuWeatherHistoricalProvider(cnf),
            out.MakeOpenWeatherHistoricalProvider(cnf),
          },
				),
				[]business.Observer{
          out.MakeCliObserver(true),
          out.MakeHttpObserver(c),
        },
			),
		},
		C: c,
	}

	cons.Consume()
}
