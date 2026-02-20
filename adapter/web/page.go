package web

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pelyib/weather-logger/domain"
)

// Page is the full data model passed to the HTML template.
type Page struct {
	Title       string
	Breadcrumbs [][]Breadcrumb
	Chart       chartJSChart
}

// Breadcrumb is a single navigation item.
type Breadcrumb struct {
	Title      string
	UriPart    string
	UriValue   string
	IsSelected bool
}

func makeBreadcrumb(title, uriPart, uriValue string, selected bool) Breadcrumb {
	return Breadcrumb{Title: title, UriPart: uriPart, UriValue: uriValue, IsSelected: selected}
}

func makeCities(currentLoc domain.Location, locations []domain.Location) []Breadcrumb {
	out := make([]Breadcrumb, 0, len(locations))
	for _, loc := range locations {
		selected := loc.Name == currentLoc.Name && loc.Country.Alpha2Code == currentLoc.Country.Alpha2Code
		out = append(out, makeBreadcrumb(
			fmt.Sprintf("%s - %s", loc.Country.Alpha2Code, loc.Name),
			"c&c",
			fmt.Sprintf("%s/%s", strings.ToLower(loc.Country.Alpha2Code), strings.ToLower(loc.Name)),
			selected,
		))
	}
	return out
}

func makeYears(ym string) []Breadcrumb {
	thisYear := time.Now().Year()
	displayedYear, _ := time.Parse("2006-01", ym)
	out := make([]Breadcrumb, 0)
	for y := 2021; y <= thisYear; y++ {
		out = append(out, makeBreadcrumb(strconv.Itoa(y), "year", strconv.Itoa(y), displayedYear.Year() == y))
	}
	return out
}

func makeMonths(ym string) []Breadcrumb {
	displayed, _ := time.Parse("2006-01", ym)
	m := displayed.Month()
	return []Breadcrumb{
		makeBreadcrumb("January", "month", "01", m == 1),
		makeBreadcrumb("February", "month", "02", m == 2),
		makeBreadcrumb("March", "month", "03", m == 3),
		makeBreadcrumb("April", "month", "04", m == 4),
		makeBreadcrumb("May", "month", "05", m == 5),
		makeBreadcrumb("June", "month", "06", m == 6),
		makeBreadcrumb("July", "month", "07", m == 7),
		makeBreadcrumb("August", "month", "08", m == 8),
		makeBreadcrumb("September", "month", "09", m == 9),
		makeBreadcrumb("October", "month", "10", m == 10),
		makeBreadcrumb("November", "month", "11", m == 11),
		makeBreadcrumb("December", "month", "12", m == 12),
	}
}

func makeBreadcrumbs(currentLoc domain.Location, ym string, locations []domain.Location) [][]Breadcrumb {
	return [][]Breadcrumb{
		{makeBreadcrumb("he!!o we4th3r", "noop", "noop", true)},
		makeCities(currentLoc, locations),
		makeYears(ym),
		makeMonths(ym),
	}
}
