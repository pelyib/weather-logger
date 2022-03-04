package shared

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Cnf struct {
	Ver string    `yaml:"version"`
	Hc  HttpCnf   `yaml:"http"`
	Lc  LoggerCnf `yaml:"logger"`
}

type Database struct {
	Folder   string   `yaml:"folder"`
	FileName string   `yaml:"fileName"`
	Buckets  []string `yaml:"buckets"`
}

type Mq struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Domain   string `yaml:"domain"`
	Port     uint16 `yaml:"port"`
	Vhost    string `yaml:"vhost"`
}

type HttpCnf struct {
	Template struct {
		Index string `yaml:"index"`
	} `yaml:"template"`
	Port     uint16   `yaml:"port"`
	Database Database `yaml:"database"`
	Mq       Mq       `yaml:"mq"`
}

type LoggerCnf struct {
	Database          Database  `yaml:"database"`
	Mq                Mq        `yaml:"mq"`
	Cities            []CityCnf `yaml:"cities"`
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
	Name        string  `yaml:"name"`
	Country     string  `yaml:"country"`
	Longitude   float64 `yaml:"longitude"`
	Langitude   float64 `yaml:"langitude"`
	Locationkey string  `yaml:"locationkey"`
}

var cnf Cnf

func load(l Logger) (*Cnf, error) {
	if len(cnf.Ver) > 0 {
		l.Info("cnf already loaded")
		return &cnf, nil
	}

	buf, err := ioutil.ReadFile(os.Getenv("CONFIG_FILE"))

	if err != nil {
		l.Error("File can not be loaded")
		return nil, err
	}
	err = yaml.Unmarshal(buf, &cnf)

	if err != nil {
		l.Error("Invalid YAML file")
		return nil, err
	}

	return &cnf, nil
}

func CreateHttpConf(l Logger) (*HttpCnf, error) {
	cnf, err := load(l)

	if err != nil {
		return nil, err
	}

	out, _ := yaml.Marshal(cnf.Hc)
	l.Info(fmt.Sprintf("Loaded cnf\n%s", string(out)))

	return &cnf.Hc, nil
}

func CreateLoggerConf(l Logger) (*LoggerCnf, error) {
	cnf, err := load(l)

	if err != nil {
		return nil, err
	}

	out, _ := yaml.Marshal(cnf.Lc)
	l.Info(fmt.Sprintf("Loaded cnf\n%s", string(out)))

	return &cnf.Lc, nil
}
