package api

import (
	"os"
	"testing"
)

func TestParseForecast(t *testing.T) {
	f, err := os.Open("testdata/data.ok.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	fc, err := parseForecast(f)
	if err != nil {
		t.Fatal(err)
	}

	ex := Forecast{
		Temperature:     9.3,
		Humidity:        83.3,
		RainProbability: 41.2,
	}

	if fc.Temperature != ex.Temperature {
		t.Fatalf("temperature is wrong; expected %f, got %f", ex.Temperature, fc.Temperature)
	}
	if fc.Humidity != ex.Humidity {
		t.Fatalf("humidity is wrong; expected %f, got %f", ex.Humidity, fc.Humidity)
	}
	if fc.RainProbability != ex.RainProbability {
		t.Fatalf("rain probability is wrong; expected %f, got %f", ex.RainProbability, fc.RainProbability)
	}
}
