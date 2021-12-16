package api

import "time"

type Weatherdata struct {
	Product Product `xml:"product"`
}

type Product struct {
	Time []Time `xml:"time"`
}

type Time struct {
	From     time.Time `xml:"from,attr"`
	To       time.Time `xml:"to,attr"`
	Location Location  `xml:"location"`
}

type Location struct {
	Altitude  int32   `xml:"altitude,attr"`
	Latitude  float32 `xml:"latitude,attr"`
	Longitude float32 `xml:"longitude,attr"`

	Temperature         Temperature         `xml:"temperature"`
	WindDirection       WindDirection       `xml:"windDirection"`
	WindSpeed           WindSpeed           `xml:"windSpeed"`
	GlobalRadiation     GlobalRadiation     `xml:"globalRadiation"`
	Humidity            Humidity            `xml:"humidity"`
	Pressure            Pressure            `xml:"pressure"`
	Cloudiness          Cloudiness          `xml:"cloudiness"`
	LowClouds           LowClouds           `xml:"lowClouds"`
	MediumClouds        MediumClouds        `xml:"mediumClouds"`
	HighClouds          HighClouds          `xml:"highClouds"`
	DewpointTemperature DewpointTemperature `xml:"dewpointTemperature"`
	Precipitation       Precipitation       `xml:"precipitation"`
	Symbol              Symbol              `xml:"symbol"`
}

func (l *Location) IsRain() bool {
	// rainfall forecast don't have temperature data
	return l.Temperature.Unit == ""
}

type Temperature struct {
	ID    string  `xml:"id,attr"`
	Unit  string  `xml:"unit,attr"`
	Value float32 `xml:"value,attr"`
}

type WindDirection struct {
	ID      string  `xml:"id,attr"`
	Degrees float32 `xml:"deg,attr"`
	Name    string  `xml:"name,attr"`
}

type WindSpeed struct {
	ID       string  `xml:"id,attr"`
	Speed    float32 `xml:"mps,attr"`
	Beaufort int     `xml:"beaufort,attr"`
	Name     string  `xml:"name,attr"`
}

type GlobalRadiation struct {
	Value float32 `xml:"value,attr"`
	Unit  string  `xml:"unit,attr"`
}

type Humidity struct {
	Value float32 `xml:"value,attr"`
	Unit  string  `xml:"unit,attr"`
}

type Pressure struct {
	ID    string  `xml:"id,attr"`
	Unit  string  `xml:"unit,attr"`
	Value float32 `xml:"value,attr"`
}

type Cloudiness struct {
	ID      string  `xml:"id,attr"`
	Percent float32 `xml:"percent,attr"`
}

type LowClouds struct {
	ID      string  `xml:"id,attr"`
	Percent float32 `xml:"percent,attr"`
}

type MediumClouds struct {
	ID      string  `xml:"id,attr"`
	Percent float32 `xml:"percent,attr"`
}

type HighClouds struct {
	ID      string  `xml:"id,attr"`
	Percent float32 `xml:"percent,attr"`
}

type DewpointTemperature struct {
	ID    string  `xml:"id,attr"`
	Unit  string  `xml:"unit,attr"`
	Value float32 `xml:"value,attr"`
}

type Precipitation struct {
	Unit        string  `xml:"unit,attr"`
	Value       float32 `xml:"value,attr"`
	MinValue    float32 `xml:"minvalue,attr"`
	MaxValue    float32 `xml:"maxvalue,attr"`
	Probability float32 `xml:"probability,attr"`
}

type Symbol struct {
	ID     string `xml:"id,attr"`
	Number int32  `xml:"number,attr"`
}
