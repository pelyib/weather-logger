package main

import (
	"log"
	"os"

	"github.com/pelyib/weather-logger/internal"
	"github.com/pelyib/weather-logger/internal/logger/business"
	"github.com/pelyib/weather-logger/internal/logger/in"
	"github.com/pelyib/weather-logger/internal/logger/out"
	"github.com/pelyib/weather-logger/internal/shared"
	"github.com/pelyib/weather-logger/internal/shared/mq"
)

func main() {
	cnf, err := shared.CreateLoggerConf(shared.MakeCliLogger(shared.App_Logger, "Config"))

	if err != nil {
		log.Fatalln(err)
	}

	dbLogger := shared.MakeCliLogger("logger", "DB")
	dbLogger.Info("loading database")
	db := internal.MakeDb(&cnf.Database, dbLogger)
	dbLogger.Info("database loaded succesfully")

	c := mq.MakeChannel(cnf.Mq, shared.MakeCliLogger(shared.App_Logger, "MQ"))

	observers := []business.Observer{
		out.MakeCliObserver(false, shared.MakeCliLogger(shared.App_Logger, "Observer.Cli")),
		out.MakeHttpObserver(c, shared.MakeCliLogger(shared.App_Logger, "Observer.Http")),
	}

	cons := mq.Consumer{
		Exchange: "http",
		Handlers: map[string]mq.Executor{
			business.COMMAND_FETCH_FORECASTS: in.MakeFetchCommandExecutor(
				business.MakeMeasurementResultProviderPool([]business.MeasurementResultProvider{
					out.MakeAccuWeatherForecastProvider(cnf, &db, shared.MakeCliLogger(shared.App_Logger, "MeasurementProvider.Accuweather.Forecast")),
					out.MakeOpenWeatherForecastProvider(cnf, shared.MakeCliLogger(shared.App_Logger, "MeasurementProvider.Operweater.Forecast")),
				}),
				observers,
			),
			business.COMMAND_FETCH_HISTORICAL: in.MakeFetchCommandExecutor(
				business.MakeMeasurementResultProviderPool(
					[]business.MeasurementResultProvider{
						out.MakeAccuWeatherHistoricalProvider(cnf, shared.MakeCliLogger(shared.App_Logger, "MeasurementProvider.Accuweather.Historical")),
						out.MakeOpenWeatherHistoricalProvider(cnf, shared.MakeCliLogger(shared.App_Logger, "MeasurementProvider.Operweater.Historical")),
					},
				),
				observers,
			),
		},
		C: c,
		L: shared.MakeCliLogger(shared.App_Logger, "MQ.consumer"),
	}

	err = cons.Consume()

	if err != nil {
		os.Exit(17)
	}
}
