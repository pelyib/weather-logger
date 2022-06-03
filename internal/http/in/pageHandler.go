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
		business.Breadcrumb{Title: "DE - Trier", Link: "/de/trier/2022/06", IsSelected: true},
		business.Breadcrumb{Title: "HU - Budapest", Link: "/hu/budapest/2022/06", IsSelected: false},
	}
	years := []business.Breadcrumb{
		business.Breadcrumb{Title: "2021", Link: "/de/trier/2021/06", IsSelected: false},
		business.Breadcrumb{Title: "2022", Link: "/de/trier/2022/06", IsSelected: true},
	}
	months := []business.Breadcrumb{
		business.Breadcrumb{"January", "/de/trier/2022/01", false},
		business.Breadcrumb{"February", "/de/trier/2022/02", false},
		business.Breadcrumb{"March", "/de/trier/2022/03", false},
		business.Breadcrumb{"April", "/de/trier/2022/04", false},
		business.Breadcrumb{"May", "/de/trier/2022/05", false},
		business.Breadcrumb{"June", "/de/trier/2022/06", true},
		business.Breadcrumb{"July", "/de/trier/2022/07", false},
		business.Breadcrumb{"August", "/de/trier/2022/08", false},
		business.Breadcrumb{"September", "/de/trier/2022/09", false},
		business.Breadcrumb{"October", "/de/trier/2022/10", false},
		business.Breadcrumb{"November", "/de/trier/2022/11", false},
		business.Breadcrumb{"December", "/de/trier/2022/12", false},
	}

	hw := business.Page{
		Title:       "he!!o we4th3r",
		Breadcrumbs: [][]business.Breadcrumb{cities, years, months},
		Chart:       chart,
	}

	log.Println(hw.Breadcrumbs)

	err = tmpl.Execute(w, hw)
	if err != nil {
		log.Println(fmt.Sprintf("template execution: %s", err))
	}
}
