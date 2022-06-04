package main

import (
	"errors"
	"io"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"
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
