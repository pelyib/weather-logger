package out

import "testing"

func TestOpenWeatherForecast_sourceId_returnsIt(t *testing.T) {
	awf := owForecast{}
	if awf.sourceId() != "openweather.forecast" {
		t.Errorf("Expected openweather.forecast, got %s", awf.sourceId())
	}
}
