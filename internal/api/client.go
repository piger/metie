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

func FetchForecast() error {
	lat := 0.0
	long := 0.0
	now := time.Now()

	url := baseURL
	subs := map[string]string{
		"${lat}":   fmt.Sprintf("%g", lat),
		"${long}":  fmt.Sprintf("%g", long),
		"${now}":   now.Format(dateTimeFormat),
		"${later}": now.Add(1 * time.Hour).Format(dateTimeFormat),
	}

	for pattern, repl := range subs {
		url = strings.Replace(url, pattern, repl, 1)
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var w Weatherdata
	if err := xml.Unmarshal(body, &w); err != nil {
		return err
	}

	for _, f := range w.Product.Time {
		fmt.Printf("From: %s, To: %s\n", f.From, f.To)
		if f.Location.IsRain() {
			fmt.Printf("Rain: %f %f %f %f\n",
				f.Location.Precipitation.Value, f.Location.Precipitation.MinValue,
				f.Location.Precipitation.MaxValue, f.Location.Precipitation.Probability)
		} else {
			fmt.Printf("temperature: %.2fC\n", f.Location.Temperature.Value)
			fmt.Printf("humidity: %.2f%%\n", f.Location.Humidity.Value)
		}
		fmt.Println()
	}

	if len(w.Product.Time) == 2 {
		f := Forecast{
			From:         w.Product.Time[0].From,
			To:           w.Product.Time[0].To,
			ForecastData: w.Product.Time[0].Location.ForecastData,
			RainfallData: w.Product.Time[1].Location.RainfallData,
		}

		fmt.Printf("%+v\n", f)
	}

	return nil
}
