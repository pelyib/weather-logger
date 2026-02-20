package domain

// Location is a geographic point to be monitored.
type Location struct {
	Name    string
	Country Country
	Lat     float64
	Lon     float64
	// AccuWeatherKey is the location identifier used by the AccuWeather API.
	AccuWeatherKey string
}

type Country struct {
	Name       string
	Alpha2Code string
}
