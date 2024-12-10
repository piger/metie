package api

import (
	"context"
	"flag"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	flagUpdate = flag.Bool("update", false, "Update the XML golden file for the client tests")
)

const goldenFileName = "testdata/data.ok.xml"

func updateGoldenFile(t *testing.T) {
	// Stephen's Green park in Dublin
	lat := 53.3375
	long := -6.2597
	url := prepareURL(lat, long)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("unexpected status code %d", resp.StatusCode)
	}

	fh, err := os.Create(goldenFileName)
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()

	if _, err := io.Copy(fh, resp.Body); err != nil {
		t.Fatal(err)
	}
}

func TestParseForecast(t *testing.T) {
	if *flagUpdate {
		updateGoldenFile(t)
	}

	f, err := os.Open(goldenFileName)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	fc, err := parseForecast(f)
	if err != nil {
		t.Fatal(err)
	}

	ex := Forecast{
		Temperature:     7.7,
		Humidity:        63.799999,
		RainProbability: 0,
	}

	if fc.Temperature != ex.Temperature {
		t.Fatalf("temperature is wrong; expected %f, got %f", ex.Temperature, fc.Temperature)
	}
	if fc.Humidity != ex.Humidity {
		t.Fatalf("humidity is wrong; expected %f, got %f", ex.Humidity, fc.Humidity)
	}
	if fc.RainProbability != ex.RainProbability {
		t.Fatalf("rain probability is wrong; expected %f, got %f", ex.RainProbability, fc.RainProbability)
	}
}
