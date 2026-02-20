package domain

import (
	"fmt"
	"time"
)

// MeasurementType distinguishes forecast data from historical observations.
type MeasurementType string

const (
	TypeForecast   MeasurementType = "forecast"
	TypeHistorical MeasurementType = "historical"
)

// Measurement is a single min/max temperature reading from a weather provider.
type Measurement struct {
	Source     string
	Type       MeasurementType
	Min        float64
	Max        float64
	At         time.Time
	RecordedAt time.Time
	Location   Location
}

// NewMeasurement constructs a validated Measurement. Returns an error if
// invariants are violated (empty source, unknown type, min > max).
func NewMeasurement(source string, mType MeasurementType, min, max float64, at time.Time, loc Location) (Measurement, error) {
	if source == "" {
		return Measurement{}, fmt.Errorf("source must not be empty")
	}
	if mType != TypeForecast && mType != TypeHistorical {
		return Measurement{}, fmt.Errorf("unknown type %q: must be %q or %q", mType, TypeForecast, TypeHistorical)
	}
	if min > max {
		return Measurement{}, fmt.Errorf("min (%v) must not exceed max (%v)", min, max)
	}
	return Measurement{
		Source:     source,
		Type:       mType,
		Min:        min,
		Max:        max,
		At:         at,
		RecordedAt: time.Now(),
		Location:   loc,
	}, nil
}
