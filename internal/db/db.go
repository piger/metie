package db

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/piger/metie/internal/api"
	"golang.org/x/net/proxy"
)

var columnNames = []string{
	"time",
	"temperature",
	"wind_deg",
	"wind_speed",
	"wind_beaufort",
	"radiation",
	"humidity",
	"pressure",
	"cloudiness",
	"clouds_low",
	"clouds_medium",
	"clouds_high",
	"dewpoint",
	"rain",
	"rain_probability",
}

func makeColumnString(names []string) string {
	return strings.Join(names, ",")
}

func makeValuesString(names []string) string {
	result := make([]string, len(names))
	for i := range names {
		result[i] = fmt.Sprintf("$%d", i+1)
	}

	return strings.Join(result, ",")
}

func WriteRow(ctx context.Context, fc *api.Forecast, dburl, table string) error {
	pgConfig, err := pgx.ParseConfig(dburl)
	if err != nil {
		return err
	}

	socksProxy := os.Getenv("SOCKS_PROXY")
	if socksProxy != "" {
		dialer, err := proxy.SOCKS5("tcp", socksProxy, nil, proxy.Direct)
		if err != nil {
			return err
		}

		if contextDialer, ok := dialer.(proxy.ContextDialer); ok {
			pgConfig.DialFunc = contextDialer.DialContext
		} else {
			return errors.New("failed type assertion into ContextDialer")
		}

		// if the DB host is a Tailscale host, redirect DNS lookups to MagicDNS.
		if strings.HasSuffix(pgConfig.Host, ".ts.net") {
			resolver := net.Resolver{
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{}
					return d.DialContext(ctx, network, "100.100.100.100")
				},
			}
			pgConfig.LookupFunc = resolver.LookupHost
		}
	}

	conn, err := pgx.ConnectConfig(ctx, pgConfig)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	columns := makeColumnString(columnNames)
	values := makeValuesString(columnNames)

	if _, err := conn.Exec(ctx,
		fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", table, columns, values),
		fc.Time,
		fc.Temperature,
		fc.WindDirection,
		fc.WindSpeedMps,
		fc.WindSpeedBeaufort,
		fc.SolarRadiation,
		fc.Humidity,
		fc.Pressure,
		fc.Cloudiness,
		fc.CloudsLow,
		fc.CloudsMedium,
		fc.CloudsHigh,
		fc.Dewpoint,
		fc.RainMm,
		fc.RainProbability,
	); err != nil {
		return err
	}

	return nil
}
