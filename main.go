package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/piger/metie/internal/api"
	"github.com/piger/metie/internal/db"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	configFile  = flag.String("config", "metie.toml", "Path to the configuration file")
	metricsAddr = flag.String("metrics", ":10333", "Address:port for metrics endpoint")
	dryrun      = flag.Bool("dryrun", false, "Run once and exit")
	showVersion = flag.Bool("version", false, "Show program's version")

	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"

	dbErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "metie_db_errors_total",
		Help: "The total number of database errors",
	})
)

func doWork(ctx context.Context, opts *Options) {
	fc, err := api.FetchForecast(ctx, opts.Latitude, opts.Longitude)
	if err != nil {
		log.Printf("error: cannot fetch forecast: %s", err)
		return
	}

	if err := db.WriteRow(ctx, fc, opts.DSN, opts.Table); err != nil {
		dbErrors.Inc()
		log.Printf("error: cannot write row to database: %s", err)
	}
}

func doMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(*metricsAddr, nil)
}

func runForever(opts *Options) error {
	go doMetrics()

	tick := time.NewTicker(time.Duration(opts.Interval))
	defer tick.Stop()

	ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer ctxCancel()

	// buffered channel used to trigger the first execution of the task
	firstRun := make(chan struct{}, 1)
	firstRun <- struct{}{}

Loop:
	for {
		select {
		case <-firstRun:
			doWork(ctx, opts)
		case <-tick.C:
			doWork(ctx, opts)
		case <-ctx.Done():
			ctxCancel()
			log.Println("signal received; exiting.")
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
