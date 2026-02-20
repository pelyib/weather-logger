package out

import (
	"time"

	"github.com/pelyib/weather-logger/domain"
)

// MeasurementRepository persists and retrieves weather measurements.
type MeasurementRepository interface {
	Save(m domain.Measurement) error
	FindByMonthAndLocation(year int, month time.Month, loc domain.Location) ([]domain.Measurement, error)
}
