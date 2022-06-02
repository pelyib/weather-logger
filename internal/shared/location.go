package shared

type City struct {
	Name        string      `yaml:"name"`
	Country     c           `yaml:"country"`
	GeoLocation geoLocation `yaml:"geoLocation"`
	Locationkey string      `yaml:"locationkey"`
}

type Location interface {
	Name() string
	Country() Country
	GeoLocation() GeoLocation
}

type Country interface {
	Name() string
	Alpha2Code() string
}

// https://en.wikipedia.org/wiki/List_of_ISO_3166_country_codes
type c struct {
	name       string `yaml:"name"`
	alpha2Code string `yaml:"code"`
}

func (c c) Name() string {
	return c.name
}

func (c c) Alpha2Code() string {
	return c.alpha2Code
}

type GeoLocation interface {
	Longitude() float64
	Langitude() float64
}

type geoLocation struct {
	longitude float64 `yaml:"longitude"`
	langitude float64 `yaml:"langitude"`
}

func (gl geoLocation) Longitude() float64 {
	return gl.longitude
}

func (gl geoLocation) Langitude() float64 {
	return gl.langitude
}
