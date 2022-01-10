package internal

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Cnf struct {
  Ver string `yaml:"version"`
  Hc HttpCnf `yaml:"http"`
  Lc LoggerCnf `yaml:"logger"`
}

type HttpCnf struct {
  Template struct {
    Index string `yaml:"index"`
  } `yaml:"template"`
}

type LoggerCnf struct {
  Database struct {
    Folder string `yaml:"folder"`
  } `yaml:"database"`
  Cities []CityCnf `yaml:"cities"`
  ForecastProviders struct {
    OpenWeather struct {
      AppId string
    } `yaml:"openWeather"`
    AccuWeather struct {
      AppId string
    } `yaml:"accuweather"`
  } `yaml:"forecast-providers"`
}

type CityCnf struct {
  Name string `yaml:"name"`
  Country string `yaml:"country"`
  Longitude float64 `yaml:"longitude"`
  Langitude float64 `yaml:"langitude"`
  Locationkey string `yaml:"locationkey"`
}

var cnf Cnf

func load() (*Cnf, error) {
  if len(cnf.Ver) > 0 {
    fmt.Println("cnf already loaded")
    return &cnf, nil
  }

	buf, err := ioutil.ReadFile(os.Getenv("CONFIG_FILE"))

  if err != nil {
		fmt.Printf("File can not be loaded\n")
		return nil, err
	}
	err = yaml.Unmarshal(buf, &cnf)

	if err != nil {
		fmt.Printf("Invalid YAML file\n")
		return nil, err
	}

  fmt.Println("loaded cnf: " + fmt.Sprint(cnf))

  return &cnf, nil
}

func CreateHttpConf() (*HttpCnf, error) {
  cnf, err := load()

  if err != nil {
    return nil, err
  }

  return &cnf.Hc, nil
}

func CreateLoggerConf() (*LoggerCnf, error) {
  cnf, err := load()

  if err != nil {
    return nil, err
  }

  return &cnf.Lc, nil
}
