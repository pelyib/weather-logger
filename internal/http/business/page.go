package business

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

type Page struct {
	Title       string
	Breadcrumbs [][]Breadcrumb
	Chart       Chart
}

type Breadcrumb struct {
	Title      string
	UriPart    string
	UriValue   string
	IsSelected bool
}

func makeCities(loc shared.Location, locations []shared.Location) []Breadcrumb {
	breadcrumbs := []Breadcrumb{}
	for _, location := range locations {
		city := location.Name
		alpha2Code := location.Country.Alpha2Code
		isSelected := loc.Name == city && loc.Country.Alpha2Code == alpha2Code
		breadcrumbs = append(
			breadcrumbs,
			MakeBreadcrumb(
				fmt.Sprintf("%s - %s", alpha2Code, city),
				"c&c",
				fmt.Sprintf("%s/%s", strings.ToLower(alpha2Code), strings.ToLower(city)),
				isSelected,
			),
		)
	}

	return breadcrumbs
}

func makeYears(ym string) []Breadcrumb {
	thisYear := time.Now().Year()
	displayedYear, _ := time.Parse("2006-01", ym)
	breadcrumbArray := []Breadcrumb{}
	for y := 2021; y <= thisYear; y++ {
		breadcrumbArray = append(breadcrumbArray, MakeBreadcrumb(strconv.Itoa(y), "year", strconv.Itoa(y), displayedYear.Year() == y))
	}

	return breadcrumbArray
}
func makeMonths(ym string) []Breadcrumb {
	diplayedDate, _ := time.Parse("2006-01", ym)
	return []Breadcrumb{
		MakeBreadcrumb("January", "month", "01", diplayedDate.Month() == 1),
		MakeBreadcrumb("February", "month", "02", diplayedDate.Month() == 2),
		MakeBreadcrumb("March", "month", "03", diplayedDate.Month() == 3),
		MakeBreadcrumb("April", "month", "04", diplayedDate.Month() == 4),
		MakeBreadcrumb("May", "month", "05", diplayedDate.Month() == 5),
		MakeBreadcrumb("June", "month", "06", diplayedDate.Month() == 6),
		MakeBreadcrumb("July", "month", "07", diplayedDate.Month() == 7),
		MakeBreadcrumb("August", "month", "08", diplayedDate.Month() == 8),
		MakeBreadcrumb("September", "month", "09", diplayedDate.Month() == 9),
		MakeBreadcrumb("October", "month", "10", diplayedDate.Month() == 10),
		MakeBreadcrumb("November", "month", "11", diplayedDate.Month() == 11),
		MakeBreadcrumb("December", "month", "12", diplayedDate.Month() == 12),
	}
}

func MakeBreadcrumb(t string, up string, uv string, isSelected bool) Breadcrumb {
	return Breadcrumb{Title: t, UriPart: up, UriValue: uv, IsSelected: isSelected}
}

func MakeBreadcrumbs(c Chart, locations []shared.Location) [][]Breadcrumb {
	return [][]Breadcrumb{
		[]Breadcrumb{
			MakeBreadcrumb("he!!o we4th3r", "noop", "noop", true),
		},
		makeCities(c.Loc, locations),
		makeYears(c.Ym),
		makeMonths(c.Ym),
	}
}
