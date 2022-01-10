package out

import (
  "encoding/json"
  "net/http"
  "net/url"
  "time"
  "fmt"
  "io"

  "github.com/pelyib/weather-logger/internal"
)

type OpenWeather struct {
  cnf *internal.LoggerCnf
}

func (s OpenWeather) Get(sr internal.SearchRequest) []internal.Forecast {
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
    return []internal.Forecast{}
  }

  req.URL.RawQuery = q.Encode()
  res, err := client.Do(req)

  if err != nil {
    fmt.Errorf("Fecthing forecasts from OpenWeather failed", err.Error())
    return []internal.Forecast{}
  }

  defer res.Body.Close()
  body, err := io.ReadAll(res.Body)

  if err != nil {
    fmt.Errorf("Reponse body reading failed", err.Error())
    return []internal.Forecast{}
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

  forecasts := make([]internal.Forecast, 0)

  for _, df := range decBody.Daily {
    at := time.Unix(df.Dt, 0)
    at, _ = time.Parse("2006-01-02", at.Format("2006-01-02"))

    fc := internal.Forecast{
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

func MakeOW(cnf *internal.LoggerCnf) ExternalProvider {
  return OpenWeather{cnf: cnf}
}
