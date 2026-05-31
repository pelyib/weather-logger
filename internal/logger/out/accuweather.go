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

	saveRawResponse(awf.db, bucketAccuWeather, searchRequest.Loc.Name, body, awf.l)

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
