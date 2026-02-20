package web

import (
	"testing"

	"github.com/pelyib/weather-logger/domain"
)

func TestMakeBreadcrumbs_Structure(t *testing.T) {
	locations := []domain.Location{
		{Name: "Budapest", Country: domain.Country{Name: "Hungary", Alpha2Code: "HU"}},
		{Name: "Vienna", Country: domain.Country{Name: "Austria", Alpha2Code: "AT"}},
	}
	current := locations[0]
	bcs := makeBreadcrumbs(current, "2024-03", locations)

	if len(bcs) != 4 {
		t.Fatalf("expected 4 breadcrumb rows, got %d", len(bcs))
	}

	// Row 0: app title (1 item)
	if len(bcs[0]) != 1 {
		t.Errorf("row 0: expected 1 item, got %d", len(bcs[0]))
	}
	if bcs[0][0].Title != "he!!o we4th3r" {
		t.Errorf("row 0 title: got %q", bcs[0][0].Title)
	}

	// Row 1: cities (2 items)
	if len(bcs[1]) != 2 {
		t.Errorf("row 1: expected 2 cities, got %d", len(bcs[1]))
	}

	// Row 2: years
	if len(bcs[2]) == 0 {
		t.Error("row 2: expected at least one year")
	}

	// Row 3: 12 months
	if len(bcs[3]) != 12 {
		t.Errorf("row 3: expected 12 months, got %d", len(bcs[3]))
	}
}

func TestMakeBreadcrumbs_SelectedCity(t *testing.T) {
	locations := []domain.Location{
		{Name: "Budapest", Country: domain.Country{Alpha2Code: "HU"}},
		{Name: "Vienna", Country: domain.Country{Alpha2Code: "AT"}},
	}
	bcs := makeBreadcrumbs(locations[1], "2024-01", locations)
	cities := bcs[1]

	if cities[0].IsSelected {
		t.Error("Budapest should not be selected")
	}
	if !cities[1].IsSelected {
		t.Error("Vienna should be selected")
	}
}

func TestMakeBreadcrumbs_SelectedMonth(t *testing.T) {
	loc := domain.Location{Name: "Budapest", Country: domain.Country{Alpha2Code: "HU"}}
	bcs := makeBreadcrumbs(loc, "2024-07", []domain.Location{loc})
	months := bcs[3]

	for i, m := range months {
		want := i == 6 // July is index 6
		if m.IsSelected != want {
			t.Errorf("month %d (%s): IsSelected=%v, want %v", i+1, m.Title, m.IsSelected, want)
		}
	}
}

func TestToChartJS_Labels(t *testing.T) {
	loc := domain.Location{Name: "Budapest"}
	c := domain.BuildChart("2024-01", loc, nil)
	cjs := toChartJS(c)

	// 31 comma-separated values in the JS array
	s := string(cjs.Labels)
	if s[0] != '[' || s[len(s)-1] != ']' {
		t.Errorf("labels not wrapped in brackets: %s", s)
	}
}

func TestToChartJS_DatasetCount(t *testing.T) {
	loc := domain.Location{Name: "Budapest"}
	c := domain.BuildChart("2024-01", loc, nil)
	cjs := toChartJS(c)

	// Expect 5 datasets: forecastMin, forecastMax, historicalMin, historicalMax, bubble
	if len(cjs.Datasets) != 5 {
		t.Errorf("expected 5 datasets, got %d", len(cjs.Datasets))
	}
}
