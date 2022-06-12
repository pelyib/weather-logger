package in

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pelyib/weather-logger/internal/http/business"
	"github.com/pelyib/weather-logger/internal/shared"
)

type PageHandler interface {
	Handle(w http.ResponseWriter, req *http.Request)
}

type pageHandler struct {
	srb searchRequestBuilder
	cnf *shared.HttpCnf
	r   business.ChartRepository
}

type searchRequestBuilder func(*http.Request) business.ChartSearchRequestI

func MakeIndexHandler(cnf *shared.HttpCnf, r *business.ChartRepository) PageHandler {
	return pageHandler{
		srb: func(r *http.Request) business.ChartSearchRequestI {
			return business.ChartSearchRequest{Ym: time.Now().Format("2006-01")}
		},
		cnf: cnf,
		r:   *r,
	}
}

func MakeHistoryHandler(cnf *shared.HttpCnf, r *business.ChartRepository) PageHandler {
	return pageHandler{
		srb: func(r *http.Request) business.ChartSearchRequestI {
			routeParams := strings.Split(r.URL.Path, "/")

			//country := routeParams[1]
			//city := routeParams[2]
			y := routeParams[3]
			m := routeParams[4]

			return business.ChartSearchRequest{Ym: y + "-" + m}
		},
		cnf: cnf,
		r:   *r,
	}
}

func (h pageHandler) Handle(w http.ResponseWriter, req *http.Request) {
	c := h.r.Load(h.srb(req))

	h.renderTmpl(w, *c)
}

func (h pageHandler) renderTmpl(
	w http.ResponseWriter,
	chart business.Chart,
) {
	tmpl, err := template.ParseFiles(h.cnf.Template.Index)
	if err != nil {
		log.Fatalf("template parsing failed: %s", err)
	}

	cities := []business.Breadcrumb{
		business.MakeBreadcrumb("DE - Trier", "c&c", "de/trier", true),
		business.MakeBreadcrumb("HU - Budapest", "c&c", "hu/budapest", false),
	}
	years := []business.Breadcrumb{
		business.MakeBreadcrumb("2021", "year", "2021", false),
		business.MakeBreadcrumb("2022", "year", "2022", true),
	}
	months := []business.Breadcrumb{
		business.MakeBreadcrumb("January", "month", "01", false),
		business.MakeBreadcrumb("February", "month", "02", false),
		business.MakeBreadcrumb("March", "month", "03", false),
		business.MakeBreadcrumb("April", "month", "04", false),
		business.MakeBreadcrumb("May", "month", "05", false),
		business.MakeBreadcrumb("June", "month", "06", true),
		business.MakeBreadcrumb("July", "month", "07", false),
		business.MakeBreadcrumb("August", "month", "08", false),
		business.MakeBreadcrumb("September", "month", "09", false),
		business.MakeBreadcrumb("October", "month", "10", false),
		business.MakeBreadcrumb("November", "month", "11", false),
		business.MakeBreadcrumb("December", "month", "12", false),
	}

	hw := business.Page{
		Title: "he!!o we4th3r",
		Breadcrumbs: [][]business.Breadcrumb{
			[]business.Breadcrumb{
				business.MakeBreadcrumb("he!!o we4th3r", "noop", "noop", true),
			},
			cities,
			years,
			months,
		},
		Chart: chart,
	}

	err = tmpl.Execute(w, hw)
	if err != nil {
		log.Println(fmt.Sprintf("template execution: %s", err))
	}
}
