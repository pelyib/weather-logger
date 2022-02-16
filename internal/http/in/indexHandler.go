package in

import (
	//	"bufio"
//	"encoding/json"
	"log"
	//  "fmt"
	"html/template"
	//	"log"
	"net/http"
	//  "os"
	//	"sort"
	"time"

	"github.com/pelyib/weather-logger/internal/http/business"
	"github.com/pelyib/weather-logger/internal/shared"
)

type Bubble struct {
  X int64 `json:"x"`
  Y float32 `json:"y"`
  Z int8 `json:"r"`
}

type Index struct {
  Hello string
  World string
  Chart business.Chart
  //  Labels []int64
//  Bubbles []Bubble
//  MinMaxLines MinMaxLines
}

type MinMaxLines struct {
  Min Line
  Max Line
}

type Line []LineDot

type LineDot struct {
  X int64 `json:"x"`
  Y float32 `json:"y"` 
}

func (line Line) Len() int {
  return len(line)
}

func (line Line) Swap(i, j int) {
  line[i], line[j] = line[j], line[i]
}

func (line Line) Less(i, j int) bool {
  return line[i].X < line[j].X
}

type indexHandler struct {
  r business.ChartRepository
}

func MakeIndexHandler(r *business.ChartRepository) indexHandler {
  return indexHandler{r: *r}
}

func (h indexHandler) Index(w http.ResponseWriter, req *http.Request) {
  c := h.r.Load(business.ChartSearchRequest{Ym: time.Now().Format("2006-01")})

  renderTmpl(w, c)

//  jc, _ := json.Marshal(c)

//  w.Write([]byte(jc))
}

/*
func index(w http.ResponseWriter, req *http.Request) {
  now := time.Now()
  labels := make([]int64, time.Date(now.Year(), now.Month() + 1, 0, 0, 0, 0, 0, time.UTC).Day())
  var fcs []internal.Forecast

  for i := range labels {
    date := time.Date(now.Year(), now.Month(), i + 1, 0, 0, 0, 0, time.UTC)
    labels[i] = date.UnixMilli()

    fn := fmt.Sprintf("./db/%s", date.Format("2006_01_02"))
    if _, err := os.Stat(fn); err == nil {
      f, _ := os.Open(fn)
      scanner := bufio.NewScanner(f)
      scanner.Split(bufio.ScanLines)

      for scanner.Scan() {
        line := scanner.Bytes()
        if len(line) > 0 {
          fs := internal.Forecast{}
          json.Unmarshal(line, &fs)
          fcs = append(fcs, fs)
        }
      }
    }
  }

  renderTmpl(w, labels, buildBubbles(fcs), buildMinMaxLine(fcs))
}

func buildBubbles(fcs []internal.Forecast) []Bubble {
  var byDate map[int64]map[float32][]internal.Forecast
  var bubbles []Bubble
  const bubbleR int8 = 2

  byDate = make(map[int64]map[float32][]internal.Forecast)

  for _, fc := range fcs {
    at, _ := time.Parse(time.RFC3339, fc.At)

    if fcsByDate, ok := byDate[at.UnixMilli()]; ok {
      if _, ok := fcsByDate[fc.Max]; ok {
        byDate[at.UnixMilli()][fc.Max] = append(byDate[at.UnixMilli()][fc.Max], fc)
      } else {
        byDate[at.UnixMilli()][fc.Max] = []internal.Forecast{fc}
      }

      if _, ok := fcsByDate[fc.Min]; ok {
        byDate[at.UnixMilli()][fc.Min] = append(byDate[at.UnixMilli()][fc.Min], fc)
      } else {
        byDate[at.UnixMilli()][fc.Min] = []internal.Forecast{fc}
      }
    } else {
      byDate[at.UnixMilli()] = make(map[float32][]internal.Forecast)
      byDate[at.UnixMilli()] = make(map[float32][]internal.Forecast)
      byDate[at.UnixMilli()][fc.Max] = []internal.Forecast{fc}
      byDate[at.UnixMilli()][fc.Min] = []internal.Forecast{fc}
    }
  }

  for at, fcsByDate := range byDate {
    for temp, fcsByTemp := range fcsByDate {
      bubbles = append(bubbles, Bubble{at, temp, bubbleR * int8(len(fcsByTemp))})
    }
  }

  return bubbles
}

func buildMinMaxLine(fcs []internal.Forecast) MinMaxLines {
  var minMaxLine MinMaxLines

  var minByDay = make(map[int64]LineDot)
  var maxByDay = make(map[int64]LineDot)

  for _, fc := range fcs {
    at, _ := time.Parse(time.RFC3339, fc.At)

    if ld, ok := minByDay[at.UnixMilli()]; (ok && ld.Y > fc.Min) || !ok {
      minByDay[at.UnixMilli()] = LineDot{at.UnixMilli(), fc.Min}
    }

    if ld, ok := maxByDay[at.UnixMilli()]; (ok && ld.Y < fc.Max) || !ok {
      maxByDay[at.UnixMilli()] = LineDot{at.UnixMilli(), fc.Max}
    }
  }

  minMaxLine.Min = make(Line, 0)
  minMaxLine.Max = make(Line, 0)

  for _, ld := range minByDay {
    minMaxLine.Min = append(minMaxLine.Min, ld)
  }

  for _, ld := range maxByDay {
    minMaxLine.Max = append(minMaxLine.Max, ld)
  }

  sort.Sort(minMaxLine.Min)
  sort.Sort(minMaxLine.Max)

  return minMaxLine
}
*/
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
    log.Fatalf("template execution: %s", err)
  }
}
