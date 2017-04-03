package main

import (
	"time"

	"github.com/datajet-io/peekaboo/config"
)

const configFilepath = "config.json" // path to config file
const servicesfilePath = "services.json"

func main() {
	config := config.Setup()

	// Start testing services
	testTicker := time.NewTicker(time.Second * time.Duration(config.Core.TestInterval))

	for _ = range testTicker.C {
		// run service tests
		for _, s := range config.Services {
			s.PerformChecks()
		}
	}
}
