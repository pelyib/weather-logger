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
	r business.ChartRepository
}

func MakeIndexHandler(r *business.ChartRepository) indexHandler {
	return indexHandler{r: *r}
}

func (h indexHandler) Index(w http.ResponseWriter, req *http.Request) {
	c := h.r.Load(business.ChartSearchRequest{Ym: time.Now().Format("2006-01")})

	renderTmpl(w, *c)
}

func renderTmpl(
	w http.ResponseWriter,
	chart business.Chart,
) {
	hc, err := shared.CreateHttpConf()
	if err != nil {
		log.Fatalln(err)
	}

	tmpl, err := template.ParseFiles(hc.Template.Index)
	if err != nil {
		log.Fatalf("template parsing failed: %s", err)
	}

	hw := Index{"he!!o", "w0rld", chart}
	err = tmpl.Execute(w, hw)
	if err != nil {
		log.Println(fmt.Sprintf("template execution: %s", err))
	}
}
