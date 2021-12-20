package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/piger/metie/internal/api"
)

var (
	lat  = flag.Float64("lat", 0.0, "Latitude")
	long = flag.Float64("long", 0.0, "Longitude")
)

func main() {
	flag.Parse()

	fc, err := api.FetchForecast(*lat, *long)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Forecast %s - %s (%s)\n", fc.From, fc.To, fc.Time)
	fmt.Printf("Temperature: %.2f °C\n", fc.Temperature)
	fmt.Printf("Humidity: %.2f%%\n", fc.Humidity)
	fmt.Printf("Wind speed: %.2f m/s\n", fc.WindSpeedMps)
}
