package shared

import (
	"testing"
	"time"
)

func TestNewMeasurementResult(t *testing.T) {
	loc := Location{}
	at := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		source  string
		mrType  string
		min     float32
		max     float32
		wantErr bool
	}{
		{
			name:    "valid forecast",
			source:  "OpenWeather",
			mrType:  MeasurementResult_Type_Forecast,
			min:     5.0,
			max:     15.0,
			wantErr: false,
		},
		{
			name:    "valid historical",
			source:  "AccuWeather",
			mrType:  MeasurementResult_Type_Historical,
			min:     -10.0,
			max:     0.0,
			wantErr: false,
		},
		{
			name:    "min equals max is valid",
			source:  "OpenWeather",
			mrType:  MeasurementResult_Type_Forecast,
			min:     7.0,
			max:     7.0,
			wantErr: false,
		},
		{
			name:    "min greater than max is invalid",
			source:  "OpenWeather",
			mrType:  MeasurementResult_Type_Forecast,
			min:     20.0,
			max:     10.0,
			wantErr: true,
		},
		{
			name:    "empty source is invalid",
			source:  "",
			mrType:  MeasurementResult_Type_Forecast,
			min:     5.0,
			max:     10.0,
			wantErr: true,
		},
		{
			name:    "unknown type is invalid",
			source:  "OpenWeather",
			mrType:  "unknown",
			min:     5.0,
			max:     10.0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr, err := NewMeasurementResult(tt.source, tt.mrType, tt.min, tt.max, at, loc)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if mr.Source != tt.source {
				t.Errorf("Source: got %q, want %q", mr.Source, tt.source)
			}
			if mr.Type != tt.mrType {
				t.Errorf("Type: got %q, want %q", mr.Type, tt.mrType)
			}
			if mr.Min != tt.min {
				t.Errorf("Min: got %v, want %v", mr.Min, tt.min)
			}
			if mr.Max != tt.max {
				t.Errorf("Max: got %v, want %v", mr.Max, tt.max)
			}
		})
	}
}

func TestMakeEmptyResults(t *testing.T) {
	mrs := MakeEmptyResults()
	if mrs == nil {
		t.Error("expected non-nil slice")
	}
	if len(mrs) != 0 {
		t.Errorf("expected empty slice, got len %d", len(mrs))
	}
}
