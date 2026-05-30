package business

import (
	"testing"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

// --- ChartSearchRequest ---

func TestChartSearchRequest_HasLoc(t *testing.T) {
	locFilled := shared.Location{Name: "Berlin"}
	locEmpty := shared.Location{}

	tests := []struct {
		name string
		csr  ChartSearchRequest
		want bool
	}{
		{"with location", ChartSearchRequest{Ym: "2024-01", Loc: locFilled}, true},
		{"empty location", ChartSearchRequest{Ym: "2024-01", Loc: locEmpty}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.csr.HasLoc(); got != tt.want {
				t.Errorf("HasLoc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChartSearchRequest_WithoutLoc(t *testing.T) {
	loc := shared.Location{Name: "Berlin"}
	csr := ChartSearchRequest{Ym: "2024-01", Loc: loc}
	without := csr.WithoutLoc()
	if without.HasLoc() {
		t.Error("WithoutLoc() should return a request with no location")
	}
	if without.GetYm() != "2024-01" {
		t.Errorf("WithoutLoc() should preserve Ym, got %q", without.GetYm())
	}
	// Original must be unchanged
	if !csr.HasLoc() {
		t.Error("WithoutLoc() must not mutate the original")
	}
}

// --- MakeEmptyChart ---

func TestMakeEmptyChart_Labels(t *testing.T) {
	csr := ChartSearchRequest{Ym: "2024-02"} // Feb 2024 is leap year: 29 days
	c := MakeEmptyChart(csr)

	if len(c.Labels) != 29 {
		t.Errorf("expected 29 labels for Feb 2024, got %d", len(c.Labels))
	}
	if c.Ym != "2024-02" {
		t.Errorf("Ym: got %q, want %q", c.Ym, "2024-02")
	}
}

func TestMakeEmptyChart_Labels_NonLeapFeb(t *testing.T) {
	csr := ChartSearchRequest{Ym: "2023-02"} // non-leap: 28 days
	c := MakeEmptyChart(csr)
	if len(c.Labels) != 28 {
		t.Errorf("expected 28 labels for Feb 2023, got %d", len(c.Labels))
	}
}

func TestMakeEmptyChart_LabelsAreMilliseconds(t *testing.T) {
	csr := ChartSearchRequest{Ym: "2024-01"}
	c := MakeEmptyChart(csr)

	// First label should be Jan 1 2024 00:00 UTC in milliseconds
	expected := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	if c.Labels[0] != expected {
		t.Errorf("Labels[0]: got %d, want %d", c.Labels[0], expected)
	}
}

func TestMakeEmptyChart_HasRequiredDatasets(t *testing.T) {
	csr := ChartSearchRequest{Ym: "2024-01"}
	c := MakeEmptyChart(csr)

	if len(c.Datasets) == 0 {
		t.Fatal("expected datasets, got none")
	}

	required := map[string]bool{
		DatasetLabelForecastMin:  false,
		DatasetLabelForecastMax:  false,
		DatasetLabelForecasts:    false,
		DatasetLabelHistoricalMin: false,
		DatasetLabelHistoricalMax: false,
	}
	for _, ds := range c.Datasets {
		required[ds.Label] = true
	}
	for label, found := range required {
		if !found {
			t.Errorf("missing dataset with label %q", label)
		}
	}
}

// --- Dataset.Push ---

func TestDataset_Push(t *testing.T) {
	ds := makeEmptyDataset(DatasetTypeLine, DatasetLabelForecastMin)

	ds.Push("01", Item{X: 100, Y: 5.0, R: 0})
	if len(ds.Data) != 1 {
		t.Errorf("expected 1 item, got %d", len(ds.Data))
	}

	// Overwrite with updated value
	ds.Push("01", Item{X: 100, Y: 3.0, R: 0})
	if ds.Data["01"].Y != 3.0 {
		t.Errorf("expected Y=3.0 after overwrite, got %v", ds.Data["01"].Y)
	}
}

// --- Chart.selectDataset (via public accessors) ---

func TestChart_ForecastMinLineDataset_CreatesOnMissing(t *testing.T) {
	c := Chart{Datasets: []*Dataset{}}
	ds := c.ForecastMinLineDataset()
	if ds == nil {
		t.Fatal("expected non-nil dataset")
	}
	if ds.Label != DatasetLabelForecastMin {
		t.Errorf("Label: got %q, want %q", ds.Label, DatasetLabelForecastMin)
	}
	if ds.Type != DatasetTypeLine {
		t.Errorf("Type: got %q, want %q", ds.Type, DatasetTypeLine)
	}
}

func TestChart_ForecastMaxBarDataset_Label(t *testing.T) {
	c := Chart{Datasets: []*Dataset{}}
	ds := c.ForecastMaxBarDataset()
	if ds.Label != DatasetLabelForecastMaxRange {
		t.Errorf("ForecastMaxBarDataset label: got %q, want %q", ds.Label, DatasetLabelForecastMaxRange)
	}
}

func TestChart_ForecastMinBarDataset_Label(t *testing.T) {
	c := Chart{Datasets: []*Dataset{}}
	ds := c.ForecastMinBarDataset()
	if ds.Label != DatasetLabelForecastMinRange {
		t.Errorf("ForecastMinBarDataset label: got %q, want %q", ds.Label, DatasetLabelForecastMinRange)
	}
}

func TestChart_ReturnsExistingDataset(t *testing.T) {
	existing := makeEmptyDataset(DatasetTypeLine, DatasetLabelForecastMin)
	existing.Push("01", Item{X: 1, Y: 42.0})
	c := Chart{Datasets: []*Dataset{existing}}

	ds := c.ForecastMinLineDataset()
	if len(ds.Data) != 1 {
		t.Errorf("expected existing dataset with 1 item, got %d", len(ds.Data))
	}
	if ds.Data["01"].Y != 42.0 {
		t.Errorf("expected Y=42.0, got %v", ds.Data["01"].Y)
	}
}

// --- Dataset factory functions ---

func TestMakeEmptyForecastMaxBarDataset_CorrectLabel(t *testing.T) {
	ds := MakeEmptyForecastMaxBarDataset()
	if ds.Label != DatasetLabelForecastMaxRange {
		t.Errorf("got label %q, want %q", ds.Label, DatasetLabelForecastMaxRange)
	}
	if ds.Type != DatasetTypeBar {
		t.Errorf("got type %q, want %q", ds.Type, DatasetTypeBar)
	}
}

func TestMakeEmptyForecastMinBarDataset_CorrectLabel(t *testing.T) {
	ds := MakeEmptyForecastMinBarDataset()
	if ds.Label != DatasetLabelForecastMinRange {
		t.Errorf("got label %q, want %q", ds.Label, DatasetLabelForecastMinRange)
	}
	if ds.Type != DatasetTypeBar {
		t.Errorf("got type %q, want %q", ds.Type, DatasetTypeBar)
	}
}
