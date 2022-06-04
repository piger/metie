package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/piger/metie/internal/api"
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
	conn, err := pgx.Connect(ctx, dburl)
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
