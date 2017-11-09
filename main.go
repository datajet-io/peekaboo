package main

import (
	"time"

	"github.com/datajet-io/peekaboo/config"
)

func main() {
	config := config.Setup()
	testTicker := time.NewTicker(time.Second * time.Duration(config.Core.TestIntervalSeconds))

	for _ = range testTicker.C {
		for _, s := range config.Services {
			s.PerformChecks()
		}
	}
}
