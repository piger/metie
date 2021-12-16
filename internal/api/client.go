package api

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	baseURL string = "http://metwdb-openaccess.ichec.ie/metno-wdb2ts/locationforecast?lat=${lat};long=${long};from=${now};to=${now}"
)

// const url = `http://metwdb-openaccess.ichec.ie/metno-wdb2ts/locationforecast?lat=${node.lat};long=${node.long};from=${now};to=${now}`;
// 2018-11-10T02:00
// 0.0,0.0

func FetchForecast() error {
	lat := 0.0
	long := 0.0
	now := time.Now()

	url := baseURL
	url = strings.Replace(url, "${lat}", fmt.Sprintf("%g", lat), 1)
	url = strings.Replace(url, "${long}", fmt.Sprintf("%g", long), 1)
	url = strings.Replace(url, "${now}", now.Format(time.RFC3339), 1)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
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

	return nil
}
