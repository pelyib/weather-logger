package out

import "testing"

func TestOpenWeatherForecast_sourceId_returnsIt(t *testing.T) {
	awf := owForecast{}
	if awf.SourceId() != "openweather.forecast" {
		t.Errorf("Expected openweather.forecast, got %s", awf.SourceId())
	}
}
