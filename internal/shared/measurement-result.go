package shared

import (
	"fmt"
	"time"
)

const MeasurementResult_Type_Forecast string = "forecast"
const MeasurementResult_Type_Historical string = "historical"

type MeasurementResult struct {
	Source     string   `json:"source"`
	Type       string   `json:"type"`
	Min        float32  `json:"min"`
	Max        float32  `json:"max"`
	At         string   `json:"at"`
	RecordedAt string   `json:"recordedAt"`
	Loc        Location `json:"loc"`
}

// NewMeasurementResult constructs a validated MeasurementResult. Returns an
// error if invariants are violated (e.g. min > max, or unknown type).
func NewMeasurementResult(source, mrType string, min, max float32, at time.Time, loc Location) (MeasurementResult, error) {
	if source == "" {
		return MeasurementResult{}, fmt.Errorf("source must not be empty")
	}
	if mrType != MeasurementResult_Type_Forecast && mrType != MeasurementResult_Type_Historical {
		return MeasurementResult{}, fmt.Errorf("unknown type %q: must be %q or %q",
			mrType, MeasurementResult_Type_Forecast, MeasurementResult_Type_Historical)
	}
	if min > max {
		return MeasurementResult{}, fmt.Errorf("min (%v) must not exceed max (%v)", min, max)
	}
	return MeasurementResult{
		Source:     source,
		Type:       mrType,
		Min:        min,
		Max:        max,
		At:         at.Format(time.RFC3339),
		RecordedAt: time.Now().Format(time.RFC3339),
		Loc:        loc,
	}, nil
}

func MakeEmptyResults() []MeasurementResult {
	return []MeasurementResult{}
}

type SearchRequest struct {
	Loc Location
}
