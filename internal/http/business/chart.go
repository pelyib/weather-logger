package business

import (
	"errors"
	"fmt"
	"time"

	"github.com/pelyib/weather-logger/internal/shared"
)

const DatasetTypeLine string = "line"
const DatasetTypeBubble string = "bubble"
const DatasetTypeBar string = "bar"

const DatasetLabelForecastMin string = "Forecast MIN"
const DatasetLabelForecastMax string = "Forecast MAX"
const DatasetLabelForecasts string = "Forecasts"
const DatasetLabelForecastMinRange string = "Forecast MIN range"
const DatasetLabelForecastMaxRange string = "Forecast MAX range"
const DatasetLabelHistoricalMin string = "Historical MIN"
const DatasetLabelHistoricalMax string = "Historical MAX"

// TODO implement Chart interface, use it everywhere [botond.pelyi]
type ChartI interface {
	Loc() shared.Location
}

type ChartSearchRequestI interface {
	GetYm() string
	HasLoc() bool
	GetLoc() shared.Location
	WithoutLoc() ChartSearchRequestI
}

type Chart struct {
	Ym       string `json:"ym"`
	Loc      shared.Location
	Labels   []int64    `json:"labels"`
	Datasets []*Dataset `json:"datasets"`
	IsNew    bool
}

type Dataset struct {
	Type  string          `json:"type"`
	Label string          `json:"label"`
	Data  map[string]Item `json:"data"`
}

type Item struct {
	X int64   `json:"x"`
	Y float32 `json:"y"`
	R int8    `json:"r"`
}

type ChartSearchRequest struct {
	Ym  string
	Loc shared.Location
}

func (csr ChartSearchRequest) GetYm() string {
	return csr.Ym
}

func (csr ChartSearchRequest) HasLoc() bool {
	return csr.Loc != shared.Location{}
}

func (csr ChartSearchRequest) GetLoc() shared.Location {
	return csr.Loc
}

// TODO csr should be immutable [botond.pelyi]
func (csr ChartSearchRequest) WithoutLoc() ChartSearchRequestI {
	copy := csr
	copy.Loc = shared.Location{}

	return copy
}

type ChartRepository interface {
	Load(csr ChartSearchRequestI) *Chart
	Save(c Chart)
}

func (c *Chart) ForecastMinLineDataset() *Dataset {
	ds, err := c.selectDataset(DatasetTypeLine, DatasetLabelForecastMin)
	if err != nil {
		c.Datasets = append(c.Datasets, MakeEmptyForecastMinLineDataset())
		return c.ForecastMinLineDataset()
	}

	return ds
}

func (c *Chart) ForecastMaxLineDataset() *Dataset {
	ds, err := c.selectDataset(DatasetTypeLine, DatasetLabelForecastMax)
	if err != nil {
		c.Datasets = append(c.Datasets, MakeEmptyForecastMaxLineDataset())
		return c.ForecastMaxLineDataset()
	}

	return ds
}

func (c *Chart) ForecastBubbleDataset() *Dataset {
	ds, err := c.selectDataset(DatasetTypeBubble, DatasetLabelForecasts)
	if err != nil {
		c.Datasets = append(c.Datasets, MakeEmptyForecastBubbleDataset())
		return c.ForecastBubbleDataset()
	}

	return ds
}

func (c *Chart) ForecastMaxBarDataset() *Dataset {
	ds, err := c.selectDataset(DatasetTypeBar, DatasetLabelForecastMaxRange)
	if err != nil {
		c.Datasets = append(c.Datasets, MakeEmptyForecastMaxBarDataset())
		return c.ForecastMaxBarDataset()
	}

	return ds
}

func (c *Chart) ForecastMinBarDataset() *Dataset {
	ds, err := c.selectDataset(DatasetTypeBar, DatasetLabelForecastMinRange)
	if err != nil {
		c.Datasets = append(c.Datasets, MakeEmptyForecastMinBarDataset())
		return c.ForecastMinBarDataset()
	}

	return ds
}

func (c *Chart) HistoricalMinLineDataset() *Dataset {
	ds, err := c.selectDataset(DatasetTypeLine, DatasetLabelHistoricalMin)
	if err != nil {
		c.Datasets = append(c.Datasets, MakeEmptyHistoricalMinDataset())
		return c.HistoricalMinLineDataset()
	}

	return ds
}

func (c *Chart) HistoricalMaxLineDataset() *Dataset {
	ds, err := c.selectDataset(DatasetTypeLine, DatasetLabelHistoricalMax)
	if err != nil {
		c.Datasets = append(c.Datasets, MakeEmptyHistoricalMaxDataset())
		return c.HistoricalMaxLineDataset()
	}

	return ds
}
func (ds *Dataset) Push(key string, i Item) {
	ds.Data[key] = i
}

func (c Chart) selectDataset(t string, l string) (*Dataset, error) {
	for i, ds := range c.Datasets {
		if ds.Type == t && ds.Label == l {
			return c.Datasets[i], nil
		}
	}

	return &Dataset{}, errors.New(fmt.Sprintf("Dataset (type: %s | label: %s) is missing", t, l))
}

func MakeEmptyChart(csr ChartSearchRequestI) Chart {
	ymt, _ := time.Parse("2006-01", csr.GetYm())
	labels := make([]int64, time.Date(ymt.Year(), ymt.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day())
	for i := range labels {
		labels[i] = time.Date(ymt.Year(), ymt.Month(), i+1, 0, 0, 0, 0, time.UTC).UnixMilli()
	}

	return Chart{
		Ym:     csr.GetYm(),
		Loc:    csr.GetLoc(),
		Labels: labels,
		Datasets: []*Dataset{
			MakeEmptyForecastMinLineDataset(),
			MakeEmptyForecastMaxLineDataset(),
			MakeEmptyForecastBubbleDataset(),
			MakeEmptyHistoricalMaxDataset(),
			MakeEmptyHistoricalMinDataset(),
		},
		IsNew: true,
	}
}

func MakeEmptyForecastMinLineDataset() *Dataset {
	return makeEmptyDataset(DatasetTypeLine, DatasetLabelForecastMin)
}

func MakeEmptyForecastMaxLineDataset() *Dataset {
	return makeEmptyDataset(DatasetTypeLine, DatasetLabelForecastMax)
}

func MakeEmptyForecastBubbleDataset() *Dataset {
	return makeEmptyDataset(DatasetTypeBubble, DatasetLabelForecasts)
}

func MakeEmptyForecastMaxBarDataset() *Dataset {
	return makeEmptyDataset(DatasetTypeBar, DatasetLabelForecasts)
}

func MakeEmptyForecastMinBarDataset() *Dataset {
	return makeEmptyDataset(DatasetTypeBar, DatasetLabelForecasts)
}

func MakeEmptyHistoricalMinDataset() *Dataset {
	return makeEmptyDataset(DatasetTypeLine, DatasetLabelHistoricalMin)
}

func MakeEmptyHistoricalMaxDataset() *Dataset {
	return makeEmptyDataset(DatasetTypeLine, DatasetLabelHistoricalMax)
}

func makeEmptyDataset(t string, l string) *Dataset {
	return &Dataset{
		Type:  t,
		Label: l,
		Data:  map[string]Item{},
	}
}
