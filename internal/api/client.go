package api

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL        = "http://metwdb-openaccess.ichec.ie/metno-wdb2ts/locationforecast?lat=${lat};long=${long};from=${now};to=${later}"
	dateTimeFormat = "2006-01-02T15:04"
)

type Forecast struct {
	Time              time.Time
	From              time.Time
	To                time.Time
	Temperature       float32
	WindDirection     float32
	WindDirectionName string
	WindSpeedMps      float32
	WindSpeedBeaufort int
	SolarRadiation    float32
	Humidity          float32
	Pressure          float32
	Cloudiness        float32
	CloudsLow         float32
	CloudsMedium      float32
	CloudsHigh        float32
	Dewpoint          float32
	RainMm            float32
	RainMin           float32
	RainMax           float32
	RainProbability   float32
}

func prepareURL(lat, long float64) string {
	url := baseURL
	now := time.Now()

	subs := map[string]string{
		"${lat}":   fmt.Sprintf("%g", lat),
		"${long}":  fmt.Sprintf("%g", long),
		"${now}":   now.Format(dateTimeFormat),
		"${later}": now.Add(1 * time.Hour).Format(dateTimeFormat),
	}

	for pattern, repl := range subs {
		url = strings.Replace(url, pattern, repl, 1)
	}

	return url
}

func FetchForecast(lat, long float64) (*Forecast, error) {
	url := prepareURL(lat, long)
	client := http.Client{
		Timeout: 1 * time.Minute,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad status code: %d (%s)", resp.StatusCode, string(body))
	}

	var w Weatherdata
	if err := xml.Unmarshal(body, &w); err != nil {
		return nil, err
	}

	if len(w.Product.Time) < 2 {
		return nil, fmt.Errorf("malformed result: missing rainfall (len = %d)", len(w.Product.Time))
	}

	root := w.Product.Time[0]
	general := w.Product.Time[0].Location.ForecastData
	rainfall := w.Product.Time[1].Location.RainfallData

	f := Forecast{
		Time:              time.Now().UTC(),
		From:              root.From,
		To:                root.To,
		Temperature:       general.Temperature.Value,
		WindDirection:     general.WindDirection.Degrees,
		WindDirectionName: general.WindDirection.Name,
		WindSpeedMps:      general.WindSpeed.Speed,
		WindSpeedBeaufort: general.WindSpeed.Beaufort,
		SolarRadiation:    general.GlobalRadiation.Value,
		Humidity:          general.Humidity.Value,
		Pressure:          general.Pressure.Value,
		Cloudiness:        general.Cloudiness.Percent,
		CloudsLow:         general.LowClouds.Percent,
		CloudsMedium:      general.MediumClouds.Percent,
		CloudsHigh:        general.HighClouds.Percent,
		Dewpoint:          general.DewpointTemperature.Value,
		RainMm:            rainfall.Precipitation.Value,
		RainMin:           rainfall.Precipitation.MinValue,
		RainMax:           rainfall.Precipitation.MaxValue,
		RainProbability:   rainfall.Precipitation.Probability,
	}

	return &f, nil
}
