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
	From time.Time
	To   time.Time
	ForecastData
	RainfallData
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

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var w Weatherdata
	if err := xml.Unmarshal(body, &w); err != nil {
		return nil, err
	}

	if len(w.Product.Time) != 2 {
		return nil, fmt.Errorf("malformed result: missing rainfall (len = %d)", len(w.Product.Time))
	}

	f := Forecast{
		From:         w.Product.Time[0].From,
		To:           w.Product.Time[0].To,
		ForecastData: w.Product.Time[0].Location.ForecastData,
		RainfallData: w.Product.Time[1].Location.RainfallData,
	}

	return &f, nil
}
