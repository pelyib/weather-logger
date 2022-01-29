package internal

type Forecast struct {
  Source string
  Min float32
  Max float32
  At string
  RecordedAt string
  // TODO: add location as well [botond.pelyi]
}

type SearchRequest struct {
}
