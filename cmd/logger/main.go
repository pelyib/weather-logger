package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/pelyib/weather-logger/internal"
	"github.com/pelyib/weather-logger/internal/logger/business"
	"github.com/pelyib/weather-logger/internal/logger/in"
	"github.com/pelyib/weather-logger/internal/logger/out"
	"github.com/pelyib/weather-logger/internal/shared"
	"github.com/pelyib/weather-logger/internal/shared/mq"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cnf, err := shared.CreateLoggerConf(shared.MakeCliLogger(shared.App_Logger, "Config"))
	if err != nil {
		log.Fatalln(err)
	}

	dbLogger := shared.MakeCliLogger("logger", "DB")
	dbLogger.Info("loading database")
	db, err := internal.MakeDb(&cnf.Database, dbLogger)
	if err != nil {
		log.Fatalln(err)
	}
	dbLogger.Info("database loaded successfully")

	c, err := mq.MakeChannel(cnf.Mq, shared.MakeCliLogger(shared.App_Logger, "MQ"))
	if err != nil {
		log.Fatalln(err)
	}

	observers := []business.Observer{
		out.MakeCliObserver(false, shared.MakeCliLogger(shared.App_Logger, "Observer.Cli")),
		out.MakeHttpObserver(c, shared.MakeCliLogger(shared.App_Logger, "Observer.Http")),
	}

	cons := mq.Consumer{
		Exchange: "http",
		Handlers: map[string]mq.Executor{
			business.COMMAND_FETCH_FORECASTS: in.MakeFetchCommandExecutor(
				business.MakeMeasurementResultProviderPool([]business.MeasurementResultProvider{
					out.MakeAccuWeatherForecastProvider(cnf, db, shared.MakeCliLogger(shared.App_Logger, "MeasurementProvider.Accuweather.Forecast")),
					out.MakeOpenWeatherForecastProvider(cnf, db, shared.MakeCliLogger(shared.App_Logger, "MeasurementProvider.Openweather.Forecast")),
					out.MakeOpenMeteoForecastProvider(cnf, db, shared.MakeCliLogger(shared.App_Logger, "MeasurementProvider.OpenMeteo.Forecast")),
				}),
				observers,
			),
			business.COMMAND_FETCH_HISTORICAL: in.MakeFetchCommandExecutor(
				business.MakeMeasurementResultProviderPool([]business.MeasurementResultProvider{
					out.MakeOpenMeteoHistoricalProvider(cnf, db, shared.MakeCliLogger(shared.App_Logger, "MeasurementProvider.OpenMeteo.Historical")),
				}),
				observers,
			),
		},
		C: c,
		L: shared.MakeCliLogger(shared.App_Logger, "MQ.consumer"),
	}

	if err := cons.Consume(ctx); err != nil {
		log.Fatalln(err)
	}
}
