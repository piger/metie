package api

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	baseURL        = "http://openaccess.pf.api.met.ie/metno-wdb2ts/locationforecast?lat=${lat};long=${long};from=${now};to=${later}"
	dateTimeFormat = "2006-01-02T15:04"
)

var (
	successes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "metie_successes_total",
		Help: "The total number of successful requests",
	})
	errors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "metie_errors_total",
			Help: "The total number of failed requests",
		},
		[]string{"reason"},
	)
	responseTimes = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "metie_response_time_seconds",
			Help:    "The response time of the server",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status_code"},
	)
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

func parseForecast(rc io.ReadCloser) (*Forecast, error) {
	var wd Weatherdata
	if err := xml.NewDecoder(rc).Decode(&wd); err != nil {
		return nil, fmt.Errorf("error decoding XML data: %w", err)
	}

	// validation
	if len(wd.Product.Time) < 2 {
		return nil, fmt.Errorf("malformed result: missing rainfall (len = %d)", len(wd.Product.Time))
	}

	root := wd.Product.Time[0]
	general := wd.Product.Time[0].Location.ForecastData
	rainfall := wd.Product.Time[1].Location.RainfallData

	fc := Forecast{
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

	return &fc, nil
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

func FetchForecast(ctx context.Context, lat, long float64) (*Forecast, error) {
	url := prepareURL(lat, long)
	client := http.Client{
		Timeout: 1 * time.Minute,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		errors.With(prometheus.Labels{"reason": "network"}).Inc()
		return nil, err
	}
	defer resp.Body.Close()

	responseTimes.With(prometheus.Labels{"status_code": strconv.Itoa(resp.StatusCode)}).Observe(float64(time.Since(start).Seconds()))

	if resp.StatusCode != 200 {
		errors.With(prometheus.Labels{"reason": "http"}).Inc()
		return nil, fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}

	fc, err := parseForecast(resp.Body)
	if err != nil {
		errors.With(prometheus.Labels{"reason": "parsing"}).Inc()
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	successes.Inc()
	return fc, nil
}
