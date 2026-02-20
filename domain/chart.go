package domain

import (
	"fmt"
	"time"
)

// LinePoint is a single value on a line dataset keyed by day-of-month.
type LinePoint struct {
	At    time.Time
	Value float64
}

// BubblePoint is a scatter point whose visual weight reflects how many
// forecast runs predicted the same temperature on the same day.
type BubblePoint struct {
	At     time.Time
	Value  float64
	Weight int // incremented for each duplicate {day, temperature} pair
}

// Chart is the visualization of measurements for a single calendar month at
// one location. It is computed from a set of Measurements, not persisted.
type Chart struct {
	YearMonth string   // "2006-01"
	Location  Location
	Labels    []time.Time // one entry per day in the month

	ForecastMinLine   map[string]LinePoint   // keyed by day "02"
	ForecastMaxLine   map[string]LinePoint
	HistoricalMinLine map[string]LinePoint
	HistoricalMaxLine map[string]LinePoint
	ForecastBubble    map[string]BubblePoint // keyed by "02_value"
}

// BuildChart aggregates measurements into a Chart for the given year-month
// string ("2006-01") and location. Measurements that do not belong to ym or
// loc are ignored.
func BuildChart(ym string, loc Location, measurements []Measurement) Chart {
	t, _ := time.Parse("2006-01", ym)
	daysInMonth := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()

	labels := make([]time.Time, daysInMonth)
	for i := range labels {
		labels[i] = time.Date(t.Year(), t.Month(), i+1, 0, 0, 0, 0, time.UTC)
	}

	c := Chart{
		YearMonth:         ym,
		Location:          loc,
		Labels:            labels,
		ForecastMinLine:   make(map[string]LinePoint),
		ForecastMaxLine:   make(map[string]LinePoint),
		HistoricalMinLine: make(map[string]LinePoint),
		HistoricalMaxLine: make(map[string]LinePoint),
		ForecastBubble:    make(map[string]BubblePoint),
	}

	for _, m := range measurements {
		day := m.At.Format("02")

		switch m.Type {
		case TypeForecast:
			if existing, ok := c.ForecastMinLine[day]; !ok || m.Min < existing.Value {
				c.ForecastMinLine[day] = LinePoint{At: m.At, Value: m.Min}
			}
			if existing, ok := c.ForecastMaxLine[day]; !ok || m.Max > existing.Value {
				c.ForecastMaxLine[day] = LinePoint{At: m.At, Value: m.Max}
			}

			minKey := fmt.Sprintf("%s_%g", day, m.Min)
			maxKey := fmt.Sprintf("%s_%g", day, m.Max)
			if existing, ok := c.ForecastBubble[minKey]; ok {
				c.ForecastBubble[minKey] = BubblePoint{At: existing.At, Value: existing.Value, Weight: existing.Weight + 2}
			} else {
				c.ForecastBubble[minKey] = BubblePoint{At: m.At, Value: m.Min, Weight: 2}
			}
			if existing, ok := c.ForecastBubble[maxKey]; ok {
				c.ForecastBubble[maxKey] = BubblePoint{At: existing.At, Value: existing.Value, Weight: existing.Weight + 2}
			} else {
				c.ForecastBubble[maxKey] = BubblePoint{At: m.At, Value: m.Max, Weight: 2}
			}

		case TypeHistorical:
			if existing, ok := c.HistoricalMinLine[day]; !ok || m.Min < existing.Value {
				c.HistoricalMinLine[day] = LinePoint{At: m.At, Value: m.Min}
			}
			if existing, ok := c.HistoricalMaxLine[day]; !ok || m.Max > existing.Value {
				c.HistoricalMaxLine[day] = LinePoint{At: m.At, Value: m.Max}
			}
		}
	}

	return c
}
