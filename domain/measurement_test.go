package domain_test

import (
	"testing"
	"time"

	"github.com/pelyib/weather-logger/domain"
)

func TestNewMeasurement(t *testing.T) {
	loc := domain.Location{Name: "Budapest"}
	at := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		source  string
		mType   domain.MeasurementType
		min     float64
		max     float64
		wantErr bool
	}{
		{name: "valid forecast", source: "OpenWeather", mType: domain.TypeForecast, min: -5, max: 10, wantErr: false},
		{name: "valid historical", source: "AccuWeather", mType: domain.TypeHistorical, min: 0, max: 0, wantErr: false},
		{name: "min equals max", source: "OpenWeather", mType: domain.TypeForecast, min: 5, max: 5, wantErr: false},
		{name: "min exceeds max", source: "OpenWeather", mType: domain.TypeForecast, min: 10, max: 5, wantErr: true},
		{name: "empty source", source: "", mType: domain.TypeForecast, min: 0, max: 10, wantErr: true},
		{name: "unknown type", source: "OpenWeather", mType: domain.MeasurementType("unknown"), min: 0, max: 10, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := domain.NewMeasurement(tc.source, tc.mType, tc.min, tc.max, at, loc)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if m.Source != tc.source {
				t.Errorf("source: got %q, want %q", m.Source, tc.source)
			}
			if m.Type != tc.mType {
				t.Errorf("type: got %q, want %q", m.Type, tc.mType)
			}
			if m.Min != tc.min {
				t.Errorf("min: got %v, want %v", m.Min, tc.min)
			}
			if m.Max != tc.max {
				t.Errorf("max: got %v, want %v", m.Max, tc.max)
			}
			if m.At != at {
				t.Errorf("at: got %v, want %v", m.At, at)
			}
			if m.RecordedAt.IsZero() {
				t.Error("recorded_at must not be zero")
			}
		})
	}
}
