package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/piger/metie/internal/api"
	"github.com/piger/metie/internal/db"
)

var (
	configFile  = flag.String("config", "metie.toml", "Path to the configuration file")
	dryrun      = flag.Bool("dryrun", false, "Run once and exit")
	showVersion = flag.Bool("version", false, "Show program's version")

	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func doWork(ctx context.Context, opts *Options) {
	fc, err := api.FetchForecast(ctx, opts.Latitude, opts.Longitude)
	if err != nil {
		log.Printf("error: cannot fetch forecast: %s", err)
		return
	}

	if err := db.WriteRow(ctx, fc, opts.DSN, opts.Table); err != nil {
		log.Printf("error: cannot write row to database: %s", err)
	}
}

func runForever(opts *Options) error {
	ctx := context.Background()
	doWork(ctx, opts)

	tick := time.NewTicker(time.Duration(opts.Interval))
	defer tick.Stop()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGTERM, syscall.SIGINT)

Loop:
	for {
		select {
		case <-tick.C:
			doWork(ctx, opts)
		case sig := <-sigch:
			log.Printf("signal received: %s", sig)
			break Loop
		}
	}

	return nil
}

func runOnce(opts *Options) error {
	ctx := context.Background()
	fc, err := api.FetchForecast(ctx, opts.Latitude, opts.Longitude)
	if err != nil {
		return err
	}

	fmt.Printf("Forecast %s - %s (%s)\n", fc.From, fc.To, fc.Time)
	fmt.Printf("Temperature: %.2f °C\n", fc.Temperature)
	fmt.Printf("Humidity: %.2f%%\n", fc.Humidity)
	fmt.Printf("Wind speed: %.2f m/s\n", fc.WindSpeedMps)
	fmt.Printf("Wind direction: %.2f %s\n", fc.WindDirection, fc.WindDirectionName)
	return nil
}

func run() error {
	flag.Parse()

	opts, err := readConfig(*configFile)
	if err != nil {
		return fmt.Errorf("error: invalid configuration: %w", err)
	}

	if *showVersion {
		if commit != "none" {
			fmt.Printf("metie %s (commit=%s, built at %s by %s)\n", version, commit, date, builtBy)
		} else {
			revision, modified, ok := buildinfo()
			if !ok {
				fmt.Printf("unknown revision\n")
			} else {
				fmt.Printf("https://github.com/piger/metie/commit/%s (modified: %v)\n", revision, modified)
			}
		}
		return nil
	}

	if *dryrun {
		return runOnce(opts)
	}

	return runForever(opts)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
