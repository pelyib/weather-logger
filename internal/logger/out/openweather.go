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
)

type OpenWeather struct {
  cnf *shared.LoggerCnf
}

func (s OpenWeather) Get(sr business.SearchRequest) []business.Forecast {
  client := http.Client{}
  q := url.Values{}
  q.Add("lat", fmt.Sprintf("%f", s.cnf.Cities[0].Langitude))
  q.Add("lon", fmt.Sprintf("%f", s.cnf.Cities[0].Longitude))
  q.Add("exclude", "current,minutely,hourly,alerts")
  q.Add("appid", s.cnf.ForecastProviders.OpenWeather.AppId)
  q.Add("units", "metric")

  req, err := http.NewRequest("GET", "https://api.openweathermap.org/data/2.5/onecall", nil)

  if (err != nil) {
    fmt.Errorf("Can not build reuqest | Reason:", err.Error())
    return []business.Forecast{}
  }

  req.URL.RawQuery = q.Encode()
  res, err := client.Do(req)

  if err != nil {
    fmt.Errorf("Fecthing forecasts from OpenWeather failed", err.Error())
    return []business.Forecast{}
  }

  defer res.Body.Close()
  body, err := io.ReadAll(res.Body)

  if err != nil {
    fmt.Errorf("Reponse body reading failed", err.Error())
    return []business.Forecast{}
  }

  var decBody struct {
    Daily []struct {
      Dt int64
      Temp struct {
        Min float32
        Max float32
      }
      Pressure uint16
    }
  }

  json.Unmarshal(body, &decBody)

  forecasts := make([]business.Forecast, 0)

  for _, df := range decBody.Daily {
    at := time.Unix(df.Dt, 0)
    at, _ = time.Parse("2006-01-02", at.Format("2006-01-02"))

    fc := business.Forecast{
      Source: "OpenWeather",
      Min: df.Temp.Min,
      Max: df.Temp.Max,
      At: at.Format(time.RFC3339),
      RecordedAt: time.Now().Format(time.RFC3339),
    }

    forecasts = append(forecasts, fc)
  }

  return forecasts
}

func MakeOW(cnf *shared.LoggerCnf) business.ForecastProvider {
  return OpenWeather{cnf: cnf}
}
