package out

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pelyib/weather-logger/internal/logger/business"
	"github.com/pelyib/weather-logger/internal/shared"
	bolt "go.etcd.io/bbolt"
)

type awForecast struct {
	cnf *shared.LoggerCnf
	db  *bolt.DB
	l   shared.Logger
	now func() time.Time
}

type awHistorical struct {
	cnf *shared.LoggerCnf
	l   shared.Logger
	now func() time.Time
}

func (awh awHistorical) sourceId() string {
	return "accuweather.historical"
}

func (awh awHistorical) fetch(sr shared.SearchRequest) ([]byte, error) {
	client := http.Client{}
	q := url.Values{}
	q.Add("apikey", awh.cnf.ForecastProviders.AccuWeather.AppId)

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/currentconditions/v1/%s/historical/24",
			awh.cnf.ForecastProviders.AccuWeather.Host,
			sr.Loc.Providers.AccuWeather.Locationkey,
		),
		nil,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not build http.request, reason: %s", err.Error()))
	}

	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Fetching Historical from Accuweather failed, reason: %s", err.Error()))
	}

	if res.StatusCode >= 400 {
		return nil, errors.New(fmt.Sprintf("Fetching Historical from Accuweather failed, HTTP status code: %s", res.Status))
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Response body reading failed, reason: %s", err.Error()))
	}

	return body, nil
}

func (awh awHistorical) mapToMeasurements(rawApiRes []byte, loc shared.Location) ([]shared.MeasurementResult, error) {
	var HistoricalDecodedResponseBody []struct {
		EpochTime   int64 `json:"EpochTime"`
		Temperature struct {
			Metric struct {
				Value    float32 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Metric"`
		} `json:"Temperature"`
	}

	err := json.Unmarshal(rawApiRes, &HistoricalDecodedResponseBody)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not unmarshal rawApiRes, reason: %s", err.Error()))
	}

	var min, max float32 = 60.0, -55.0
	today, _ := time.Parse("2006/01/02", awh.now().Format("2006/01/02"))
	todayUnixMilli := today.Unix()

	for _, i := range HistoricalDecodedResponseBody {
		if i.EpochTime < todayUnixMilli {
			continue
		}

		if i.Temperature.Metric.Value < min {
			min = i.Temperature.Metric.Value
		}

		if i.Temperature.Metric.Value > max {
			max = i.Temperature.Metric.Value
		}
	}

	mrs := shared.MakeEmptyResults()
	mrs = append(
		mrs,
		shared.MeasurementResult{
			Source:     "AccuWeather",
			Type:       shared.MeasurementResult_Type_Historical,
			Min:        min,
			Max:        max,
			At:         today.Add(time.Hour * 24 * -1).Format(time.RFC3339),
			RecordedAt: awh.now().Format(time.RFC3339),
			Loc:        loc,
		},
	)

	return mrs, nil
}

func (awf awForecast) sourceId() string {
	return "accuweather.forecast"
}

func (awf awForecast) fetch(sr shared.SearchRequest) ([]byte, error) {
	client := http.Client{}
	q := url.Values{}
	q.Add("apikey", awf.cnf.ForecastProviders.AccuWeather.AppId)
	q.Add("metric", "true")

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/forecasts/v1/daily/5day/%s",
			awf.cnf.ForecastProviders.AccuWeather.Host,
			sr.Loc.Providers.AccuWeather.Locationkey,
		),
		nil,
	)

	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Fetching Forecasts from Accuweather failed, reassson: %s", err.Error()))
	}

	if res.StatusCode >= 400 {
		return nil, errors.New(fmt.Sprintf("Fetching Forecasts from Accuweather failed, HTTP status code: %s", res.Status))
	}

	defer res.Body.Close()
	remoteApiResponse, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Response body reading failed, reason: %s", err.Error()))
	}

	return remoteApiResponse, nil
}

func (awf awForecast) mapToMeasurements(rawApiRes []byte, loc shared.Location) ([]shared.MeasurementResult, error) {
	mrs := shared.MakeEmptyResults()
	var decBody struct {
		DailyForecasts []struct {
			Date        string
			Temperature struct {
				Minimum struct {
					Value float32
				}
				Maximum struct {
					Value float32
				}
			}
		}
	}

	err := json.Unmarshal(rawApiRes, &decBody)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not unmarshal rawApiRes, reason: %s", err.Error()))
	}

	for _, df := range decBody.DailyForecasts {
		at, _ := time.Parse(time.RFC3339, df.Date)
		at, _ = time.Parse("2006-01-02", at.Format("2006-01-02"))

		mrs = append(
			mrs,
			shared.MeasurementResult{
				Source:     "AccuWeather",
				Type:       shared.MeasurementResult_Type_Forecast,
				Min:        df.Temperature.Minimum.Value,
				Max:        df.Temperature.Maximum.Value,
				At:         at.Format(time.RFC3339),
				RecordedAt: awf.now().Format(time.RFC3339),
				Loc:        loc,
			},
		)
	}

	return mrs, nil
}

func MakeAccuWeatherForecastProvider(cnf *shared.LoggerCnf, db *bolt.DB, l shared.Logger) business.MeasurementResultProvider {
	return Fetcher{
		weatherProviderAdapter: awForecast{cnf: cnf, db: db, l: l, now: now},
		dbClient:               MakeNewClient(cnf.CouchDb),
		logger:                 l,
	}
}

func MakeAccuWeatherHistoricalProvider(cnf *shared.LoggerCnf, l shared.Logger) business.MeasurementResultProvider {
	return Fetcher{
		weatherProviderAdapter: awHistorical{cnf: cnf, l: l, now: now},
		dbClient:               MakeNewClient(cnf.CouchDb),
		logger:                 l,
	}
}
