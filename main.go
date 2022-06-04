package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/piger/metie/internal/api"
	"github.com/piger/metie/internal/db"
)

var (
	lat         = flag.Float64("lat", 0.0, "Latitude")
	long        = flag.Float64("long", 0.0, "Longitude")
	dbconfig    = flag.String("dbconfig", "", "Path to a file containing a DB URI")
	showVersion = flag.Bool("version", false, "Show program's version")

	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
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
		fmt.Printf("Wind direction: %.2f %s\n", fc.WindDirection, fc.WindDirectionName)
	}

	return nil
}

func main() {
	flag.Parse()

	if *lat == 0.0 || *long == 0.0 {
		fmt.Fprintln(os.Stderr, "error: you must specify both -lat and -long")
		os.Exit(1)
	}
	if *showVersion {
		fmt.Printf("metie %s (commit=%s, built at %s by %s)\n", version, commit, date, builtBy)
		return
	}

	if err := run(); err != nil {
		log.Fatalf("ERROR: %s", err)
	}
}
