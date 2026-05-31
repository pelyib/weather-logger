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
	db     *bolt.DB
}

func (owf owForecast) GetMeasurement(sr shared.SearchRequest) []shared.MeasurementResult {
	q := url.Values{}
	q.Add("lat", fmt.Sprintf("%f", sr.Loc.GeoLocation.Latitude))
	q.Add("lon", fmt.Sprintf("%f", sr.Loc.GeoLocation.Longitude))
	q.Add("appid", owf.cnf.ForecastProviders.OpenWeather.AppId)
	q.Add("units", "metric")

	req, err := http.NewRequest("GET", "https://api.openweathermap.org/data/2.5/forecast", nil)
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

	saveRawResponse(owf.db, bucketOpenWeather, sr.Loc.Name, body, owf.l)

	var decBody struct {
		List []struct {
			Dt   int64 `json:"dt"`
			Main struct {
				TempMin float32 `json:"temp_min"`
				TempMax float32 `json:"temp_max"`
			} `json:"main"`
		} `json:"list"`
	}

	if err := json.Unmarshal(body, &decBody); err != nil {
		owf.l.Error(fmt.Sprintf("Could not parse response body, reason: %s", err.Error()))
		return shared.MakeEmptyResults()
	}

	type dayMinMax struct {
		min float32
		max float32
	}
	dayMap := map[string]*dayMinMax{}
	dayOrder := []string{}

	for _, entry := range decBody.List {
		day := time.Unix(entry.Dt, 0).UTC().Format("2006-01-02")
		if d, ok := dayMap[day]; ok {
			if entry.Main.TempMin < d.min {
				d.min = entry.Main.TempMin
			}
			if entry.Main.TempMax > d.max {
				d.max = entry.Main.TempMax
			}
		} else {
			dayMap[day] = &dayMinMax{min: entry.Main.TempMin, max: entry.Main.TempMax}
			dayOrder = append(dayOrder, day)
		}
	}

	mrs := shared.MakeEmptyResults()

	for _, day := range dayOrder {
		at, _ := time.Parse("2006-01-02", day)
		d := dayMap[day]
		mrs = append(mrs, shared.MeasurementResult{
			Source:     "OpenWeather",
			Type:       shared.MeasurementResult_Type_Forecast,
			Min:        d.min,
			Max:        d.max,
			At:         at.Format(time.RFC3339),
			RecordedAt: time.Now().Format(time.RFC3339),
			Loc:        sr.Loc,
		})
	}

	return mrs
}

func MakeOpenWeatherForecastProvider(cnf *shared.LoggerCnf, db *bolt.DB, l shared.Logger) business.MeasurementResultProvider {
	return owForecast{cnf: cnf, l: l, client: &http.Client{}, db: db}
}
