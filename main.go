package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
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
	reqSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "metie_requests_success_total",
		Help: "The total number of successful requests",
	})
	reqError = promauto.NewCounter(prometheus.CounterOpts{
		Name: "metie_requests_error_total",
		Help: "The total number of failed requests",
	})
)

func doWork(ctx context.Context, opts *Options) {
	fc, err := api.FetchForecast(ctx, opts.Latitude, opts.Longitude)
	if err != nil {
		reqError.Inc()
		slog.Error("failed fetching forecast", "err", err)
		return
	}

	if err := db.WriteRow(ctx, fc, opts.DSN, opts.Table); err != nil {
		dbErrors.Inc()
		slog.Error("failed writing to database", "err", err)
		return
	}
	reqSuccess.Inc()
}

func doMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(*metricsAddr, nil)
}

func runForever(ctx context.Context, opts *Options) error {
	go doMetrics()

	timer := time.NewTimer(1 * time.Millisecond)
	defer timer.Stop()

	ctx, ctxCancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer ctxCancel()

Loop:
	for {
		select {
		case <-timer.C:
			slog.Debug("fetching forecast")
			doWork(ctx, opts)
			timer.Reset(time.Duration(opts.Interval))
		case <-ctx.Done():
			ctxCancel()
			slog.Info("signal received; exiting")
			break Loop
		}
	}

	return nil
}

func runOnce(ctx context.Context, opts *Options) error {
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
		return fmt.Errorf("invalid configuration: %w", err)
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

	ctx := context.Background()

	if *dryrun {
		return runOnce(ctx, opts)
	}

	return runForever(ctx, opts)
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("error", "err", err)
		os.Exit(1)
	}
}
