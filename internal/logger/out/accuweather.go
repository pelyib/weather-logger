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
	cnf    *shared.LoggerCnf
	db     *bolt.DB
	l      shared.Logger
	client *http.Client
}

type awHistorical struct {
	cnf    *shared.LoggerCnf
	l      shared.Logger
	client *http.Client
}

func (awh awHistorical) GetMeasurement(searchRequest shared.SearchRequest) []shared.MeasurementResult {
	mrs := shared.MakeEmptyResults()

	q := url.Values{}
	q.Add("metric", "true")

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"http://dataservice.accuweather.com/currentconditions/v1/%s/historical/24",
			searchRequest.Loc.Providers.AccuWeather.Locationkey,
		),
		nil,
	)
	if err != nil {
		awh.l.Error(fmt.Sprintf("Could not build http.request, reason: %s", err.Error()))
		return mrs
	}

	req.URL.RawQuery = q.Encode()
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", awh.cnf.ForecastProviders.AccuWeather.AppId))

	res, err := awh.client.Do(req)
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

	var historicalBody []struct {
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

	if err := json.Unmarshal(body, &historicalBody); err != nil {
		awh.l.Error(fmt.Sprintf("Could not parse response body, reason: %s", err.Error()))
		return mrs
	}

	var min, max float32 = 60.0, -55.0
	today, _ := time.Parse("2006/01/02", time.Now().Format("2006/01/02"))
	todayUnixMilli := today.Unix()

	for _, i := range historicalBody {
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
			Loc:        searchRequest.Loc,
		},
	)

	return mrs
}

func (awf awForecast) GetMeasurement(searchRequest shared.SearchRequest) []shared.MeasurementResult {
	mrs := shared.MakeEmptyResults()

	q := url.Values{}
	q.Add("metric", "true")

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"https://dataservice.accuweather.com/forecasts/v1/daily/5day/%s",
			searchRequest.Loc.Providers.AccuWeather.Locationkey,
		),
		nil,
	)
	if err != nil {
		awf.l.Error(fmt.Sprintf("Could not build request, reason: %s", err.Error()))
		return mrs
	}

	req.URL.RawQuery = q.Encode()
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", awf.cnf.ForecastProviders.AccuWeather.AppId))

	res, err := awf.client.Do(req)
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

	if err := awf.db.Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte("accuweather.raw_response"))
		if b == nil {
			return fmt.Errorf("bucket accuweather.raw_response not found")
		}
		if err := b.Put([]byte(time.Now().Format(time.UnixDate)), body); err != nil {
			return err
		}
		awf.l.Info("Put done")
		return nil
	}); err != nil {
		awf.l.Error(fmt.Sprintf("Failed to save raw response, reason: %s", err.Error()))
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

	if err := json.Unmarshal(body, &decBody); err != nil {
		awf.l.Error(fmt.Sprintf("Could not parse response body, reason: %s", err.Error()))
		return mrs
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
				RecordedAt: time.Now().Format(time.RFC3339),
				Loc:        searchRequest.Loc,
			},
		)
	}

	return mrs
}

func MakeAccuWeatherForecastProvider(cnf *shared.LoggerCnf, db *bolt.DB, l shared.Logger) business.MeasurementResultProvider {
	return awForecast{cnf: cnf, db: db, l: l, client: &http.Client{}}
}

func MakeAccuWeatherHistoricalProvider(cnf *shared.LoggerCnf, l shared.Logger) business.MeasurementResultProvider {
	return awHistorical{cnf: cnf, l: l, client: &http.Client{}}
}
