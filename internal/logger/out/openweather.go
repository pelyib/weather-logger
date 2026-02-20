package out

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pelyib/weather-logger/internal/logger/business"
	"github.com/pelyib/weather-logger/internal/shared"
)

type owHourlyResponse struct {
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
	Timezone       string  `json:"timezone"`
	TimezoneOffset int     `json:"timezone_offset"`
	Hourly         []struct {
		Dt        int64   `json:"dt"`
		Temp      float32 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
		Weather   []struct {
			ID          int    `json:"id"`
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		Rain struct {
			OneH float64 `json:"1h"`
		} `json:"rain,omitempty"`
	} `json:"hourly"`
}

type owForecast struct {
	cnf    *shared.LoggerCnf
	l      shared.Logger
	client *http.Client
}

type owHistorical struct {
	cnf    *shared.LoggerCnf
	l      shared.Logger
	client *http.Client
}

func (owh owHistorical) GetMeasurement(sr shared.SearchRequest) []shared.MeasurementResult {
	mrs := shared.MakeEmptyResults()
	today, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	yesterday := today.Add(time.Hour * 24 * -1)

	q := url.Values{}
	q.Add("lat", fmt.Sprintf("%f", sr.Loc.GeoLocation.Latitude))
	q.Add("lon", fmt.Sprintf("%f", sr.Loc.GeoLocation.Longitude))
	q.Add("appid", owh.cnf.ForecastProviders.OpenWeather.AppId)
	q.Add("units", "metric")
	q.Add("dt", strconv.FormatInt(yesterday.Unix(), 10))

	req, err := http.NewRequest("GET", "https://api.openweathermap.org/data/2.5/onecall/timemachine", nil)
	if err != nil {
		owh.l.Error(fmt.Sprintf("Can not build request, reason: %s", err.Error()))
		return mrs
	}

	req.URL.RawQuery = q.Encode()
	res, err := owh.client.Do(req)
	if err != nil {
		owh.l.Error(fmt.Sprintf("Fetching forecasts from OpenWeather failed, reason: %s", err.Error()))
		return mrs
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		owh.l.Error(fmt.Sprintf("Response body reading failed, reason: %s", err.Error()))
		return mrs
	}

	var decBody owHourlyResponse
	if err := json.Unmarshal(body, &decBody); err != nil {
		owh.l.Error(fmt.Sprintf("Could not parse response body, reason: %s", err.Error()))
		return mrs
	}

	var min, max float32 = 60.0, -60.0
	todayUnix := today.Unix()
	yesterdayUnix := yesterday.Unix()

	for _, i := range decBody.Hourly {
		if i.Dt > todayUnix || i.Dt < yesterdayUnix {
			continue
		}
		if i.Temp > max {
			max = i.Temp
		}
		if i.Temp < min {
			min = i.Temp
		}
	}

	mrs = append(
		mrs,
		shared.MeasurementResult{
			Source:     "OpenWeather",
			Type:       shared.MeasurementResult_Type_Historical,
			Min:        min,
			Max:        max,
			At:         yesterday.Format(time.RFC3339),
			RecordedAt: time.Now().Format(time.RFC3339),
			Loc:        sr.Loc,
		},
	)

	return mrs
}

func (owf owForecast) GetMeasurement(sr shared.SearchRequest) []shared.MeasurementResult {
	q := url.Values{}
	q.Add("lat", fmt.Sprintf("%f", sr.Loc.GeoLocation.Latitude))
	q.Add("lon", fmt.Sprintf("%f", sr.Loc.GeoLocation.Longitude))
	q.Add("exclude", "current,minutely,hourly,alerts")
	q.Add("appid", owf.cnf.ForecastProviders.OpenWeather.AppId)
	q.Add("units", "metric")

	req, err := http.NewRequest("GET", "https://api.openweathermap.org/data/2.5/onecall", nil)
	if err != nil {
		owf.l.Error(fmt.Sprintf("Can not build request, reason: %s", err.Error()))
		return shared.MakeEmptyResults()
	}

	req.URL.RawQuery = q.Encode()
	res, err := owf.client.Do(req)
	if err != nil {
		owf.l.Error(fmt.Sprintf("Fetching forecasts from OpenWeather failed, reason: %s", err.Error()))
		return shared.MakeEmptyResults()
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		owf.l.Error(fmt.Sprintf("Response body reading failed, reason: %s", err.Error()))
		return shared.MakeEmptyResults()
	}

	var decBody struct {
		Daily []struct {
			Dt   int64
			Temp struct {
				Min float32
				Max float32
			}
			Pressure uint16
		}
	}

	if err := json.Unmarshal(body, &decBody); err != nil {
		owf.l.Error(fmt.Sprintf("Could not parse response body, reason: %s", err.Error()))
		return shared.MakeEmptyResults()
	}

	mrs := shared.MakeEmptyResults()

	for _, df := range decBody.Daily {
		at := time.Unix(df.Dt, 0)
		at, _ = time.Parse("2006-01-02", at.Format("2006-01-02"))

		mrs = append(mrs, shared.MeasurementResult{
			Source:     "OpenWeather",
			Type:       shared.MeasurementResult_Type_Forecast,
			Min:        df.Temp.Min,
			Max:        df.Temp.Max,
			At:         at.Format(time.RFC3339),
			RecordedAt: time.Now().Format(time.RFC3339),
			Loc:        sr.Loc,
		})
	}

	return mrs
}

func MakeOpenWeatherForecastProvider(cnf *shared.LoggerCnf, l shared.Logger) business.MeasurementResultProvider {
	return owForecast{cnf: cnf, l: l, client: &http.Client{}}
}

func MakeOpenWeatherHistoricalProvider(cnf *shared.LoggerCnf, l shared.Logger) business.MeasurementResultProvider {
	return owHistorical{cnf: cnf, l: l, client: &http.Client{}}
}
