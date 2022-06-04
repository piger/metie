package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pelletier/go-toml/v2"
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

// Duration is a type alias used by the TOML parser.
type Duration time.Duration

func (d *Duration) UnmarshalText(b []byte) error {
	x, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = Duration(x)
	return nil
}

type Options struct {
	Latitude  float64
	Longitude float64
	DSN       string
	Table     string
	Interval  Duration
}

func readConfig(filename string) (*Options, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	data, err := io.ReadAll(fh)
	if err != nil {
		return nil, err
	}

	var opts Options
	if err := toml.Unmarshal(data, &opts); err != nil {
		return nil, err
	}

	if opts.Latitude == 0.0 || opts.Longitude == 0.0 {
		return nil, errors.New("latitude and longitude are missing or invalid")
	} else if opts.DSN == "" || opts.Table == "" {
		return nil, errors.New("missing DSN or table name")
	} else if opts.Interval == 0 {
		return nil, errors.New("invalid interval")
	}

	return &opts, nil
}

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

func runI(opts *Options) error {
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
	opts, err := readConfig(*configFile)
	if err != nil {
		return fmt.Errorf("invalid configuration: %s", err)
	}

	if *dryrun {
		return runOnce(opts)
	}

	return runI(opts)
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("metie %s (commit=%s, built at %s by %s)\n", version, commit, date, builtBy)
		return
	}

	if err := run(); err != nil {
		log.Fatalf("error: %s", err)
	}
}
