package business

import (
	"testing"

	"github.com/pelyib/weather-logger/internal/shared"
)

func TestMakeBreadcrumb(t *testing.T) {
	bc := MakeBreadcrumb("Berlin", "city", "berlin", true)
	if bc.Title != "Berlin" {
		t.Errorf("Title: got %q, want %q", bc.Title, "Berlin")
	}
	if bc.UriPart != "city" {
		t.Errorf("UriPart: got %q, want %q", bc.UriPart, "city")
	}
	if bc.UriValue != "berlin" {
		t.Errorf("UriValue: got %q, want %q", bc.UriValue, "berlin")
	}
	if !bc.IsSelected {
		t.Error("IsSelected should be true")
	}
}

func TestMakeBreadcrumbs_Structure(t *testing.T) {
	loc := shared.Location{Name: "Berlin"}
	loc.Country.Alpha2Code = "DE"

	locations := []shared.Location{loc}
	chart := Chart{Ym: "2024-01", Loc: loc}

	bcs := MakeBreadcrumbs(chart, locations)

	// Should have 4 rows: app title, cities, years, months
	if len(bcs) != 4 {
		t.Errorf("expected 4 breadcrumb rows, got %d", len(bcs))
	}

	// Row 0: app title — always 1 entry
	if len(bcs[0]) != 1 {
		t.Errorf("expected 1 app-title breadcrumb, got %d", len(bcs[0]))
	}

	// Row 1: cities — matches locations list
	if len(bcs[1]) != len(locations) {
		t.Errorf("expected %d city breadcrumbs, got %d", len(locations), len(bcs[1]))
	}

	// Row 2: years — at least one entry
	if len(bcs[2]) == 0 {
		t.Error("expected at least one year breadcrumb")
	}

	// Row 3: months — 12 entries
	if len(bcs[3]) != 12 {
		t.Errorf("expected 12 month breadcrumbs, got %d", len(bcs[3]))
	}
}

func TestMakeBreadcrumbs_SelectedCity(t *testing.T) {
	berlin := shared.Location{Name: "Berlin"}
	berlin.Country.Alpha2Code = "DE"
	paris := shared.Location{Name: "Paris"}
	paris.Country.Alpha2Code = "FR"

	chart := Chart{Ym: "2024-01", Loc: berlin}
	bcs := MakeBreadcrumbs(chart, []shared.Location{berlin, paris})

	var selectedCount int
	for _, bc := range bcs[1] {
		if bc.IsSelected {
			selectedCount++
		}
	}
	if selectedCount != 1 {
		t.Errorf("expected exactly 1 selected city, got %d", selectedCount)
	}
	if bcs[1][0].IsSelected != true {
		t.Error("Berlin should be selected (first city)")
	}
	if bcs[1][1].IsSelected != false {
		t.Error("Paris should not be selected")
	}
}

func TestMakeBreadcrumbs_SelectedMonth(t *testing.T) {
	loc := shared.Location{}
	chart := Chart{Ym: "2024-03", Loc: loc}
	bcs := MakeBreadcrumbs(chart, []shared.Location{loc})

	months := bcs[3]
	// March is index 2 (0-based)
	if !months[2].IsSelected {
		t.Error("March (index 2) should be selected for ym=2024-03")
	}
	for i, m := range months {
		if i != 2 && m.IsSelected {
			t.Errorf("month at index %d should not be selected", i)
		}
	}
}
