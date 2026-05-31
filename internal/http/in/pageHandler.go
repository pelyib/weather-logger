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
	srb  searchRequestBuilder
	cnf  *shared.HttpCnf
	r    business.ChartRepository
	tmpl *template.Template
}

type searchRequestBuilder func([]shared.Location, *http.Request) business.ChartSearchRequestI

func MakeIndexHandler(cnf *shared.HttpCnf, r *business.ChartRepository) (PageHandler, error) {
	tmpl, err := template.ParseFiles(cnf.Template.Index)
	if err != nil {
		return nil, err
	}
	return pageHandler{
		srb: func(locations []shared.Location, r *http.Request) business.ChartSearchRequestI {
			return business.ChartSearchRequest{Ym: time.Now().Format("2006-01"), Loc: locations[0]}
		},
		cnf:  cnf,
		r:    *r,
		tmpl: tmpl,
	}, nil
}

func MakeHistoryHandler(cnf *shared.HttpCnf, r *business.ChartRepository) (PageHandler, error) {
	tmpl, err := template.ParseFiles(cnf.Template.Index)
	if err != nil {
		return nil, err
	}
	return pageHandler{
		srb: func(locations []shared.Location, r *http.Request) business.ChartSearchRequestI {
			routeParams := strings.Split(r.URL.Path, "/")
			var loc shared.Location
			country := routeParams[1]
			city := routeParams[2]
			for _, location := range locations {
				if strings.ToLower(location.Name) == strings.ToLower(city) && location.Country.Alpha2Code == strings.ToUpper(country) {
					loc = location
					continue
				}
			}
			// TODO: not found in case of missing location [botond.pelyi]

			y := routeParams[3]
			m := routeParams[4]

			return business.ChartSearchRequest{Ym: y + "-" + m, Loc: loc}
		},
		cnf:  cnf,
		r:    *r,
		tmpl: tmpl,
	}, nil
}

func (h pageHandler) Handle(w http.ResponseWriter, req *http.Request) {
	c := h.r.Load(h.srb(h.cnf.Locations, req))

	h.renderTmpl(w, *c)
}

func (h pageHandler) renderTmpl(
	w http.ResponseWriter,
	chart business.Chart,
) {
	hw := business.Page{
		Title:       "he!!o we4th3r",
		Breadcrumbs: business.MakeBreadcrumbs(chart, h.cnf.Locations),
		Chart:       chart,
	}

	if err := h.tmpl.Execute(w, hw); err != nil {
		log.Println(fmt.Sprintf("template execution: %s", err))
	}
}
