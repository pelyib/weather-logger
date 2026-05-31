package out

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pelyib/weather-logger/internal/logger/business"
	"github.com/pelyib/weather-logger/internal/shared"
)

type omForecast struct {
	cnf    *shared.LoggerCnf
	l      shared.Logger
	client *http.Client
}

func (omf omForecast) GetMeasurement(sr shared.SearchRequest) []shared.MeasurementResult {
	models := omf.cnf.ForecastProviders.OpenMeteo.Models
	if len(models) == 0 {
		return shared.MakeEmptyResults()
	}

	q := url.Values{}
	q.Add("latitude", fmt.Sprintf("%f", sr.Loc.GeoLocation.Latitude))
	q.Add("longitude", fmt.Sprintf("%f", sr.Loc.GeoLocation.Longitude))
	q.Add("daily", "temperature_2m_max,temperature_2m_min")
	q.Add("models", strings.Join(models, ","))
	q.Add("timezone", omf.cnf.ForecastProviders.OpenMeteo.Timezone)
	q.Add("timeformat", "unixtime")

	req, err := http.NewRequest("GET", "https://api.open-meteo.com/v1/forecast", nil)
	if err != nil {
		omf.l.Error(fmt.Sprintf("Can not build request, reason: %s", err.Error()))
		return shared.MakeEmptyResults()
	}

	req.URL.RawQuery = q.Encode()
	res, err := omf.client.Do(req)
	if err != nil {
		omf.l.Error(fmt.Sprintf("Fetching forecasts from Open-Meteo failed, reason: %s", err.Error()))
		return shared.MakeEmptyResults()
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		omf.l.Error(fmt.Sprintf("Response body reading failed, reason: %s", err.Error()))
		return shared.MakeEmptyResults()
	}

	var decBody struct {
		Daily map[string]json.RawMessage `json:"daily"`
	}
	if err := json.Unmarshal(body, &decBody); err != nil {
		omf.l.Error(fmt.Sprintf("Could not parse response body, reason: %s", err.Error()))
		return shared.MakeEmptyResults()
	}

	var timestamps []int64
	if err := json.Unmarshal(decBody.Daily["time"], &timestamps); err != nil {
		omf.l.Error(fmt.Sprintf("Could not parse daily timestamps, reason: %s", err.Error()))
		return shared.MakeEmptyResults()
	}

	mrs := shared.MakeEmptyResults()

	for _, model := range models {
		rawMax, hasMax := decBody.Daily["temperature_2m_max_"+model]
		rawMin, hasMin := decBody.Daily["temperature_2m_min_"+model]
		if !hasMax || !hasMin {
			omf.l.Warning(fmt.Sprintf("Model %s not found in response, skipping", model))
			continue
		}

		var maxVals []*float32
		var minVals []*float32
		if err := json.Unmarshal(rawMax, &maxVals); err != nil {
			omf.l.Warning(fmt.Sprintf("Could not parse max values for model %s, skipping", model))
			continue
		}
		if err := json.Unmarshal(rawMin, &minVals); err != nil {
			omf.l.Warning(fmt.Sprintf("Could not parse min values for model %s, skipping", model))
			continue
		}

		for i, ts := range timestamps {
			if i >= len(maxVals) || i >= len(minVals) {
				break
			}
			if maxVals[i] == nil || minVals[i] == nil {
				continue
			}
			at := time.Unix(ts, 0).UTC()
			mrs = append(mrs, shared.MeasurementResult{
				Source:     "OpenMeteo/" + model,
				Type:       shared.MeasurementResult_Type_Forecast,
				Min:        *minVals[i],
				Max:        *maxVals[i],
				At:         at.Format(time.RFC3339),
				RecordedAt: time.Now().Format(time.RFC3339),
				Loc:        sr.Loc,
			})
		}
	}

	return mrs
}

func MakeOpenMeteoForecastProvider(cnf *shared.LoggerCnf, l shared.Logger) business.MeasurementResultProvider {
	return omForecast{cnf: cnf, l: l, client: &http.Client{}}
}

type omHistorical struct {
	cnf    *shared.LoggerCnf
	l      shared.Logger
	client *http.Client
}

func (omh omHistorical) GetMeasurement(sr shared.SearchRequest) []shared.MeasurementResult {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	q := url.Values{}
	q.Add("latitude", fmt.Sprintf("%f", sr.Loc.GeoLocation.Latitude))
	q.Add("longitude", fmt.Sprintf("%f", sr.Loc.GeoLocation.Longitude))
	q.Add("start_date", yesterday)
	q.Add("end_date", yesterday)
	q.Add("daily", "temperature_2m_max,temperature_2m_min")
	q.Add("timezone", omh.cnf.ForecastProviders.OpenMeteo.Timezone)
	q.Add("timeformat", "unixtime")

	req, err := http.NewRequest("GET", "https://archive-api.open-meteo.com/v1/archive", nil)
	if err != nil {
		omh.l.Error(fmt.Sprintf("Can not build request, reason: %s", err.Error()))
		return shared.MakeEmptyResults()
	}

	req.URL.RawQuery = q.Encode()
	res, err := omh.client.Do(req)
	if err != nil {
		omh.l.Error(fmt.Sprintf("Fetching historical from Open-Meteo failed, reason: %s", err.Error()))
		return shared.MakeEmptyResults()
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		omh.l.Error(fmt.Sprintf("Response body reading failed, reason: %s", err.Error()))
		return shared.MakeEmptyResults()
	}

	var decBody struct {
		Daily struct {
			Time []int64   `json:"time"`
			Max  []float32 `json:"temperature_2m_max"`
			Min  []float32 `json:"temperature_2m_min"`
		} `json:"daily"`
	}
	if err := json.Unmarshal(body, &decBody); err != nil {
		omh.l.Error(fmt.Sprintf("Could not parse response body, reason: %s", err.Error()))
		return shared.MakeEmptyResults()
	}

	mrs := shared.MakeEmptyResults()

	for i, ts := range decBody.Daily.Time {
		if i >= len(decBody.Daily.Max) || i >= len(decBody.Daily.Min) {
			break
		}
		at := time.Unix(ts, 0).UTC()
		mrs = append(mrs, shared.MeasurementResult{
			Source:     "OpenMeteo",
			Type:       shared.MeasurementResult_Type_Historical,
			Min:        decBody.Daily.Min[i],
			Max:        decBody.Daily.Max[i],
			At:         at.Format(time.RFC3339),
			RecordedAt: time.Now().Format(time.RFC3339),
			Loc:        sr.Loc,
		})
	}

	return mrs
}

func MakeOpenMeteoHistoricalProvider(cnf *shared.LoggerCnf, l shared.Logger) business.MeasurementResultProvider {
	return omHistorical{cnf: cnf, l: l, client: &http.Client{}}
}
