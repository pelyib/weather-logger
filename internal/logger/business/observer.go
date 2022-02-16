package business

type Observer interface {
  Notify(forecasts []Forecast)
}
