package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Config is the application configuration loaded from the YAML file at
// $CONFIG_FILE. It is treated as a plain value — no global singletons.
type Config struct {
	Version  string `yaml:"version"`
	Port     uint16 `yaml:"port"`
	Template struct {
		Index string `yaml:"index"`
	} `yaml:"template"`
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`
	Scheduler struct {
		ForecastsInterval  string `yaml:"forecasts_interval"`
		HistoricalInterval string `yaml:"historical_interval"`
	} `yaml:"scheduler"`
	Providers struct {
		OpenWeather struct {
			APIKey string `yaml:"api_key"`
		} `yaml:"open_weather"`
		AccuWeather struct {
			APIKey string `yaml:"api_key"`
		} `yaml:"accu_weather"`
	} `yaml:"providers"`
	Locations []Location `yaml:"locations"`
}

// Location is the config representation of a monitored geographic point.
type Location struct {
	Name    string `yaml:"name"`
	Country struct {
		Name       string `yaml:"name"`
		Alpha2Code string `yaml:"alpha2_code"`
	} `yaml:"country"`
	Lat            float64 `yaml:"lat"`
	Lon            float64 `yaml:"lon"`
	AccuWeatherKey string  `yaml:"accu_weather_key"`
}

// Load reads and parses the YAML file at $CONFIG_FILE.
func Load() (*Config, error) {
	path := os.Getenv("CONFIG_FILE")
	if path == "" {
		return nil, fmt.Errorf("config: CONFIG_FILE env var not set")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %q: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("config: parse YAML: %w", err)
	}
	return &c, nil
}

// ForecastsInterval parses the scheduler.forecasts_interval duration.
// Defaults to 6 hours if unset or unparseable.
func (c *Config) ForecastsInterval() time.Duration {
	return parseDuration(c.Scheduler.ForecastsInterval, 6*time.Hour)
}

// HistoricalInterval parses the scheduler.historical_interval duration.
// Defaults to 24 hours if unset or unparseable.
func (c *Config) HistoricalInterval() time.Duration {
	return parseDuration(c.Scheduler.HistoricalInterval, 24*time.Hour)
}

func parseDuration(s string, fallback time.Duration) time.Duration {
	if s == "" {
		return fallback
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return fallback
	}
	return d
}
