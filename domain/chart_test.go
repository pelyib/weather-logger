package domain_test

import (
	"testing"
	"time"

	"github.com/pelyib/weather-logger/domain"
)

func TestBuildChart_Labels(t *testing.T) {
	loc := domain.Location{Name: "Test"}

	t.Run("january has 31 labels", func(t *testing.T) {
		c := domain.BuildChart("2024-01", loc, nil)
		if len(c.Labels) != 31 {
			t.Fatalf("expected 31 labels, got %d", len(c.Labels))
		}
		if c.Labels[0].Day() != 1 {
			t.Errorf("first label day: got %d, want 1", c.Labels[0].Day())
		}
		if c.Labels[30].Day() != 31 {
			t.Errorf("last label day: got %d, want 31", c.Labels[30].Day())
		}
	})

	t.Run("february 2024 (leap year) has 29 labels", func(t *testing.T) {
		c := domain.BuildChart("2024-02", loc, nil)
		if len(c.Labels) != 29 {
			t.Fatalf("expected 29 labels, got %d", len(c.Labels))
		}
	})

	t.Run("february 2023 (non-leap) has 28 labels", func(t *testing.T) {
		c := domain.BuildChart("2023-02", loc, nil)
		if len(c.Labels) != 28 {
			t.Fatalf("expected 28 labels, got %d", len(c.Labels))
		}
	})
}

func TestBuildChart_ForecastLines(t *testing.T) {
	loc := domain.Location{Name: "Budapest"}
	at := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	m1 := mustMeasurement(domain.NewMeasurement("OpenWeather", domain.TypeForecast, -5, 10, at, loc))
	// second measurement for the same day with lower min and higher max
	m2 := mustMeasurement(domain.NewMeasurement("AccuWeather", domain.TypeForecast, -8, 12, at, loc))
	// third measurement for the same day with middle values (should not override min/max)
	m3 := mustMeasurement(domain.NewMeasurement("OpenWeather", domain.TypeForecast, -3, 9, at, loc))

	c := domain.BuildChart("2024-01", loc, []domain.Measurement{m1, m2, m3})

	if pt, ok := c.ForecastMinLine["15"]; !ok {
		t.Fatal("expected entry for day 15 in ForecastMinLine")
	} else if pt.Value != -8 {
		t.Errorf("ForecastMinLine[15].Value: got %v, want -8", pt.Value)
	}

	if pt, ok := c.ForecastMaxLine["15"]; !ok {
		t.Fatal("expected entry for day 15 in ForecastMaxLine")
	} else if pt.Value != 12 {
		t.Errorf("ForecastMaxLine[15].Value: got %v, want 12", pt.Value)
	}
}

func TestBuildChart_HistoricalLines(t *testing.T) {
	loc := domain.Location{Name: "Budapest"}
	at := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)

	m := mustMeasurement(domain.NewMeasurement("OpenWeather", domain.TypeHistorical, 2, 8, at, loc))
	c := domain.BuildChart("2024-01", loc, []domain.Measurement{m})

	if pt, ok := c.HistoricalMinLine["10"]; !ok {
		t.Fatal("expected entry for day 10 in HistoricalMinLine")
	} else if pt.Value != 2 {
		t.Errorf("got %v, want 2", pt.Value)
	}

	if pt, ok := c.HistoricalMaxLine["10"]; !ok {
		t.Fatal("expected entry for day 10 in HistoricalMaxLine")
	} else if pt.Value != 8 {
		t.Errorf("got %v, want 8", pt.Value)
	}
}

func TestBuildChart_BubbleWeight(t *testing.T) {
	loc := domain.Location{Name: "Budapest"}
	at := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

	m1 := mustMeasurement(domain.NewMeasurement("OpenWeather", domain.TypeForecast, 5, 15, at, loc))
	m2 := mustMeasurement(domain.NewMeasurement("AccuWeather", domain.TypeForecast, 5, 15, at, loc))

	c := domain.BuildChart("2024-01", loc, []domain.Measurement{m1, m2})

	// Same temperature on same day → weight should be 4 (2+2)
	minKey := "05_5"
	if bp, ok := c.ForecastBubble[minKey]; !ok {
		t.Fatalf("expected bubble for key %q", minKey)
	} else if bp.Weight != 4 {
		t.Errorf("bubble weight: got %d, want 4", bp.Weight)
	}
}

func mustMeasurement(m domain.Measurement, err error) domain.Measurement {
	if err != nil {
		panic(err)
	}
	return m
}
