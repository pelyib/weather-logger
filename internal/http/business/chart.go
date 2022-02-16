package business

import (
	"errors"
	"time"
)

const DatasetTypeLine string = "line"
const DatasetTypeBubble string = "bubble"

const DatasetLabelFirecastMin string = "MIN"
const DatasetLabelFirecastMax string = "MAX"
const DatasetLabelForecasts string = "Forecasts"

type Chart struct {
  Ym string `json:"ym"`
  Labels []int64 `json:"labels"`
  Datasets []Dataset `json:"datasets"`
}

type Dataset struct {
  Type string `json:"type"`
  Label string `json:"label"`
  Data map[string]Item `json:"data"`
}

type Item struct {
  X int64 `json:"x"`
  Y float32 `json:"y"`
  R int8 `json:"r"`
}

type ChartSearchRequest struct {
  Ym string
}

type ChartRepository interface {
  Load(searchReq ChartSearchRequest) Chart
  Save(c Chart)
}

func (c Chart) MinLineDataset() (Dataset, error) {
  for _, ds := range c.Datasets {
    if ds.Type == DatasetTypeLine && ds.Label == DatasetLabelFirecastMin {
      return ds, nil
    }
  }

  return Dataset{}, errors.New("MinLineDataset not found")
}

func (c Chart) MaxLineDataset() (Dataset, error) {
  for _, ds := range c.Datasets {
    if ds.Type == DatasetTypeLine && ds.Label == DatasetLabelFirecastMax {
      return ds, nil
    }
  }

  return Dataset{}, errors.New("MaxLineDataset not found")
}

func (c Chart) ForecastBubbleDataset() (Dataset, error) {
  for _, ds := range c.Datasets {
    if ds.Type == DatasetTypeBubble && ds.Label == DatasetLabelForecasts {
      return ds, nil
    }
  }

  return Dataset{}, errors.New("ForecastBubbleDataset not found")
}

func MakeEmptyChart(Ym string) Chart {
  ymt, _ := time.Parse("2006-01", Ym)
  labels := make([]int64, time.Date(ymt.Year(), ymt.Month() +1, 0, 0, 0, 0, 0, time.UTC).Day())
  for i := range labels {
    labels[i] = time.Date(ymt.Year(), ymt.Month(), i + 1, 0, 0, 0, 0, time.UTC).UnixMilli()
  }

  return Chart{
    Ym: Ym,
    Labels: labels,
    Datasets: []Dataset{
      Dataset{Type: DatasetTypeLine, Label: DatasetLabelFirecastMin, Data: map[string]Item{},},
      Dataset{Type: DatasetTypeLine, Label: DatasetLabelFirecastMax, Data: map[string]Item{},},
      Dataset{Type: DatasetTypeBubble, Label: DatasetLabelForecasts, Data: map[string]Item{},},
    },
  }
}
