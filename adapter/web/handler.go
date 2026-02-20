package web

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/pelyib/weather-logger/domain"
	portout "github.com/pelyib/weather-logger/port/out"
)

// Handler serves the weather chart pages.
type Handler struct {
	repo          portout.MeasurementRepository
	locations     []domain.Location
	templatePath  string
	log           *slog.Logger
}

// NewHandler creates an HTTP handler wired to the given repository.
func NewHandler(repo portout.MeasurementRepository, locations []domain.Location, templatePath string, log *slog.Logger) *Handler {
	return &Handler{
		repo:         repo,
		locations:    locations,
		templatePath: templatePath,
		log:          log,
	}
}

// Index serves the current month's chart for the first configured location.
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	if len(h.locations) == 0 {
		http.Error(w, "no locations configured", http.StatusInternalServerError)
		return
	}
	ym := time.Now().Format("2006-01")
	h.render(w, ym, h.locations[0])
}

// History serves the chart for the year/month and location encoded in the URL.
// Expected URL pattern: /{alpha2}/{city}/{year}/{month}
func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}
	country := strings.ToUpper(parts[0])
	city := parts[1]
	year := parts[2]
	month := parts[3]

	var loc domain.Location
	for _, l := range h.locations {
		if strings.ToLower(l.Name) == strings.ToLower(city) && l.Country.Alpha2Code == country {
			loc = l
			break
		}
	}
	if loc.Name == "" {
		http.NotFound(w, r)
		return
	}

	ym := year + "-" + month
	h.render(w, ym, loc)
}

func (h *Handler) render(w http.ResponseWriter, ym string, loc domain.Location) {
	t, _ := time.Parse("2006-01", ym)
	measurements, err := h.repo.FindByMonthAndLocation(t.Year(), t.Month(), loc)
	if err != nil {
		h.log.Error("handler: load measurements", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	chart := domain.BuildChart(ym, loc, measurements)
	page := Page{
		Title:       "he!!o we4th3r",
		Breadcrumbs: makeBreadcrumbs(loc, ym, h.locations),
		Chart:       toChartJS(chart),
	}

	tmpl, err := template.ParseFiles(h.templatePath)
	if err != nil {
		h.log.Error("handler: parse template", "err", err)
		http.Error(w, fmt.Sprintf("template error: %v", err), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, page); err != nil {
		h.log.Error("handler: execute template", "err", err)
	}
}
