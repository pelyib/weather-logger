package out

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pelyib/weather-logger/internal"
)

type AccuWeather struct {
  cnf *internal.LoggerCnf
}

func (s AccuWeather) Get(sr internal.SearchRequest) []internal.Forecast {
  fcs := []internal.Forecast{}

  client := http.Client{}
  q := url.Values{}
  q.Add("apikey", s.cnf.ForecastProviders.AccuWeather.AppId)
  q.Add("metric", "true")

  req, err := http.NewRequest(
    "GET",
    fmt.Sprintf(
      "https://dataservice.accuweather.com/forecasts/v1/daily/5day/%s",
      s.cnf.Cities[0].Locationkey,
    ),
    nil,
  )

  if err != nil {
    fmt.Errorf("Can not build request | Reason: ", err.Error())
    return fcs
  }

  req.URL.RawQuery = q.Encode()

  res, err := client.Do(req)

  
  if err != nil {
    fmt.Errorf("Fetching Forecasts from Accuweather failed", err.Error())
    return fcs
  }

  defer res.Body.Close()
  body, err := io.ReadAll(res.Body)

  if err != nil {
    fmt.Errorf("Response body reading failed", err.Error())
    return fcs
  }

  var decBody struct {
    DailyForecasts []struct {
      Date string
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

    fcs = append(
      fcs,
      internal.Forecast{
        Source: "AccuWeather",
        Min: df.Temperature.Minimum.Value,
        Max: df.Temperature.Maximum.Value,
        At: at.Format(time.RFC3339),
        RecordedAt: time.Now().Format(time.RFC3339),
      },
    )
  }

  return fcs
}

func MakeAW(cnf *internal.LoggerCnf) ExternalProvider {
  return AccuWeather{cnf: cnf}
}
