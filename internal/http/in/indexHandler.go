package in

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/pelyib/weather-logger/internal/http/business"
	"github.com/pelyib/weather-logger/internal/shared"
)

type Bubble struct {
	X int64   `json:"x"`
	Y float32 `json:"y"`
	Z int8    `json:"r"`
}

type Index struct {
	Hello string
	World string
	Chart business.Chart
}

type indexHandler struct {
	cnf *shared.HttpCnf
	r   business.ChartRepository
}

func MakeIndexHandler(cnf *shared.HttpCnf, r *business.ChartRepository) indexHandler {
	return indexHandler{cnf: cnf, r: *r}
}

func (h indexHandler) Index(w http.ResponseWriter, req *http.Request) {
	c := h.r.Load(business.ChartSearchRequest{Ym: time.Now().Format("2006-01")})

	h.renderTmpl(w, *c)
}

func (h indexHandler) renderTmpl(
	w http.ResponseWriter,
	chart business.Chart,
) {
	tmpl, err := template.ParseFiles(h.cnf.Template.Index)
	if err != nil {
		log.Fatalf("template parsing failed: %s", err)
	}

	hw := Index{"he!!o", "w0rld", chart}
	err = tmpl.Execute(w, hw)
	if err != nil {
		log.Println(fmt.Sprintf("template execution: %s", err))
	}
}
