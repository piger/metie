package main

import (
	"log"

	"github.com/piger/metie/internal/api"
)

func main() {
	if err := api.FetchForecast(); err != nil {
		log.Println(err)
	}
}
