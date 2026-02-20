package out

import (
	"context"

	"github.com/pelyib/weather-logger/domain"
)

// Provider fetches weather measurements from an external API for a given
// location and measurement type.
type Provider interface {
	Fetch(ctx context.Context, mType domain.MeasurementType, loc domain.Location) ([]domain.Measurement, error)
}
