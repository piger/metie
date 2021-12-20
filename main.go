package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/piger/metie/internal/api"
	"github.com/piger/metie/internal/db"
)

var (
	lat      = flag.Float64("lat", 0.0, "Latitude")
	long     = flag.Float64("long", 0.0, "Longitude")
	dbconfig = flag.String("dbconfig", "", "Path to a file containing a DB URI")
)

func run() error {
	fc, err := api.FetchForecast(*lat, *long)
	if err != nil {
		return fmt.Errorf("cannot fetch forecast: %s", err)
	}

	if *dbconfig != "" {
		config, err := db.ReadConfig(*dbconfig)
		if err != nil {
			return fmt.Errorf("cannot read db config file: %s", err)
		}

		ctx := context.Background()
		if err := db.WriteRow(ctx, fc, config, "weather_forecast"); err != nil {
			return fmt.Errorf("cannot write to DB: %s", err)
		}
	} else {
		fmt.Printf("Forecast %s - %s (%s)\n", fc.From, fc.To, fc.Time)
		fmt.Printf("Temperature: %.2f °C\n", fc.Temperature)
		fmt.Printf("Humidity: %.2f%%\n", fc.Humidity)
		fmt.Printf("Wind speed: %.2f m/s\n", fc.WindSpeedMps)
	}

	return nil
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatalf("ERROR: %s", err)
	}
}
