package business

type Forecast struct {
  Source string `json:"source"`
  Min float32 `json:"min"`
  Max float32 `json:"max"`
  At string `json:"at"`
  RecordedAt string `json:"recordedAt"`
  // TODO: add location as well [botond.pelyi]
}

type SearchRequest struct {
}
