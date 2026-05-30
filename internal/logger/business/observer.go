package business

import "github.com/pelyib/weather-logger/internal/shared"

type Observer interface {
	Notify(measurementResults []shared.MeasurementResult)
}
