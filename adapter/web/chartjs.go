package web

import (
	"html/template"
	"sort"
	"strconv"
	"strings"

	"github.com/pelyib/weather-logger/domain"
)

const (
	labelForecastMin    = "Forecast MIN"
	labelForecastMax    = "Forecast MAX"
	labelForecasts      = "Forecasts"
	labelHistoricalMin  = "Historical MIN"
	labelHistoricalMax  = "Historical MAX"
	datasetTypeLine     = "line"
	datasetTypeBubble   = "bubble"
)

// chartJSDataset is the Chart.js dataset representation passed to the template.
type chartJSDataset struct {
	Type  string
	Label string
	Data  []chartJSItem
}

// chartJSItem is a single data point in a Chart.js dataset.
type chartJSItem struct {
	X int64   // milliseconds since epoch
	Y float64
	R int8    // bubble radius (0 for line datasets)
}

// chartJSChart is the full chart structure passed to the HTML template.
type chartJSChart struct {
	// Labels is a template.JS value so it renders as a proper JS array in
	// the <script> context without HTML escaping.
	Labels   template.JS
	Datasets []chartJSDataset
}

// toChartJS converts a domain.Chart to the Chart.js-ready structure used by
// the HTML template. Line datasets are sorted by day; bubble datasets are
// sorted by X then Y for deterministic rendering.
func toChartJS(c domain.Chart) chartJSChart {
	// Build Labels JS array: [ms, ms, ...] (one per day in month)
	labelParts := make([]string, len(c.Labels))
	for i, t := range c.Labels {
		labelParts[i] = strconv.FormatInt(t.UnixMilli(), 10)
	}
	labelsJS := template.JS("[" + strings.Join(labelParts, ",") + "]")

	datasets := []chartJSDataset{
		{Type: datasetTypeLine, Label: labelForecastMin, Data: lineItems(c.ForecastMinLine)},
		{Type: datasetTypeLine, Label: labelForecastMax, Data: lineItems(c.ForecastMaxLine)},
		{Type: datasetTypeLine, Label: labelHistoricalMin, Data: lineItems(c.HistoricalMinLine)},
		{Type: datasetTypeLine, Label: labelHistoricalMax, Data: lineItems(c.HistoricalMaxLine)},
		{Type: datasetTypeBubble, Label: labelForecasts, Data: bubbleItems(c.ForecastBubble)},
	}

	return chartJSChart{Labels: labelsJS, Datasets: datasets}
}

func lineItems(pts map[string]domain.LinePoint) []chartJSItem {
	items := make([]chartJSItem, 0, len(pts))
	for _, p := range pts {
		items = append(items, chartJSItem{X: p.At.UnixMilli(), Y: p.Value})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].X < items[j].X })
	return items
}

func bubbleItems(pts map[string]domain.BubblePoint) []chartJSItem {
	items := make([]chartJSItem, 0, len(pts))
	for _, p := range pts {
		r := int8(p.Weight)
		if r > 127 {
			r = 127
		}
		items = append(items, chartJSItem{X: p.At.UnixMilli(), Y: p.Value, R: r})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].X != items[j].X {
			return items[i].X < items[j].X
		}
		return items[i].Y < items[j].Y
	})
	return items
}
