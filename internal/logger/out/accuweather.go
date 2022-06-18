package out

import (
	"encoding/json"
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
}

type awHistorical struct {
	cnf *shared.LoggerCnf
	l   shared.Logger
}

func (awh awHistorical) GetMeasurement(searchRequest business.SearchRequest) []shared.MeasurementResult {
	mrs := shared.MakeEmptyResults()

	client := http.Client{}
	q := url.Values{}
	q.Add("apikey", awh.cnf.ForecastProviders.AccuWeather.AppId)

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"http://dataservice.accuweather.com/currentconditions/v1/%s/historical/24",
			awh.cnf.Cities[0].Locationkey,
		),
		nil,
	)

	if err != nil {
		awh.l.Error(fmt.Sprintf("Could not build http.request, reason: %s", err.Error()))
		return mrs
	}

	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)

	if err != nil {
		awh.l.Error(fmt.Sprintf("Fetching Forecasts from Accuweather failed, reason: %s", err.Error()))
		return mrs
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		awh.l.Error(fmt.Sprintf("Response body reading failed, reason: %s", err.Error()))
		return mrs
	}

	var HistoricalDecodedResponseBody []struct {
		LocalObservationDateTime time.Time   `json:"LocalObservationDateTime"`
		EpochTime                int64       `json:"EpochTime"`
		WeatherText              string      `json:"WeatherText"`
		WeatherIcon              int         `json:"WeatherIcon"`
		HasPrecipitation         bool        `json:"HasPrecipitation"`
		PrecipitationType        interface{} `json:"PrecipitationType"`
		IsDayTime                bool        `json:"IsDayTime"`
		Temperature              struct {
			Metric struct {
				Value    float32 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Metric"`
			Imperial struct {
				Value    int    `json:"Value"`
				Unit     string `json:"Unit"`
				UnitType int    `json:"UnitType"`
			} `json:"Imperial"`
		} `json:"Temperature"`
		MobileLink string `json:"MobileLink"`
		Link       string `json:"Link"`
	}

	json.Unmarshal(body, &HistoricalDecodedResponseBody)

	var min, max float32 = 60.0, -55.0
	today, _ := time.Parse("2006/01/02", time.Now().Format("2006/01/02"))
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

	mrs = append(
		mrs,
		shared.MeasurementResult{
			Source:     "AccuWeather",
			Type:       shared.MeasurementResult_Type_Historical,
			Min:        min,
			Max:        max,
			At:         today.Add(time.Hour * 24 * -1).Format(time.RFC3339),
			RecordedAt: time.Now().Format(time.RFC3339),
		},
	)

	return mrs
}

func (awf awForecast) GetMeasurement(searchRequest business.SearchRequest) []shared.MeasurementResult {
	mrs := shared.MakeEmptyResults()

	client := http.Client{}
	q := url.Values{}
	q.Add("apikey", awf.cnf.ForecastProviders.AccuWeather.AppId)
	q.Add("metric", "true")

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"https://dataservice.accuweather.com/forecasts/v1/daily/5day/%s",
			awf.cnf.Cities[0].Locationkey,
		),
		nil,
	)

	if err != nil {
		awf.l.Error(fmt.Sprintf("Could not build request, reason: %s", err.Error()))
		return mrs
	}

	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)

	if err != nil {
		awf.l.Error(fmt.Sprintf("Fetching Forecasts from Accuweather failed, reason: %s", err.Error()))
		return mrs
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		awf.l.Error(fmt.Sprintf("Response body reading failed, reason: %s", err.Error()))
		return mrs
	}

	err = awf.db.Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte("accuweather.raw_response"))

		err = b.Put([]byte(time.Now().Format(time.UnixDate)), body)
		awf.l.Info("Put done")
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		awf.l.Error(fmt.Sprintf("Failed to connect to DB, could not save, reason: %s", err.Error()))
	}

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

	json.Unmarshal(body, &decBody)

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
				RecordedAt: time.Now().Format(time.RFC3339),
			},
		)
	}

	return mrs
}

func MakeAccuWeatherForecastProvider(cnf *shared.LoggerCnf, db *bolt.DB, l shared.Logger) business.MeasurementResultProvider {
	return awForecast{cnf: cnf, db: db, l: l}
}

func MakeAccuWeatherHistoricalProvider(cnf *shared.LoggerCnf, l shared.Logger) business.MeasurementResultProvider {
	return awHistorical{cnf: cnf, l: l}
}
