package shared

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

func MakeEmptyResults() []MeasurementResult {
	return []MeasurementResult{}
}

type SearchRequest struct {
	Loc Location
}
