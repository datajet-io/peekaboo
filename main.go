package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cenk/backoff"
	"github.com/datajet-io/peekaboo/alerting"
	"github.com/datajet-io/peekaboo/config"
	"github.com/datajet-io/peekaboo/services"
	"github.com/uber-go/zap"
)

const configFilepath = "config.json" // path to config file
const servicesfilePath = "services.json"

//checks if there is Internet connection by pinging Google
func hasInternet(logger zap.Logger) bool {
	notify := func(err error, duration time.Duration) {
		logger.Warn(
			"Encountered error during run",
			zap.Int64("duration", duration.Nanoseconds()/int64(time.Millisecond)),
			zap.Error(err),
		)
	}

	operation := func() error {
		response, err := http.Get("https://www.google.com")

		if err != nil {
			// Pass this error back, the notify handler will log it for us
			return err
		}

		defer response.Body.Close()
		io.Copy(ioutil.Discard, response.Body)

		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			// Pass this error back, the notify handler will log it for us
			return errors.New("Returned status code was not acceptable")
		}
		return err
	}

	err := backoff.RetryNotify(operation, backoff.NewExponentialBackOff(), notify)
	return err == nil
}

func welcomeOwners(srvs *services.Services, msgs *alerting.Messaging) {
	const seperator = "|"

	// create a unique list of owners
	l := make(map[string]int, 5)

	for _, s := range srvs.Services {
		for _, owner := range s.Owners {
			l[owner.Cell+seperator+owner.Name]++
		}
	}

	// message each owner
	for k, numberOfServices := range l {
		s := strings.Split(k, seperator)
		t := "services"

		if numberOfServices == 1 {
			t = "service"
		}

		m := fmt.Sprintf("Hi %s! You will receive alerts for %d %s on Peekaboo - the messaging-based health checker. Stop alerts by texting 'stop <ID>', start alerts with 'start <ID>'", s[1], numberOfServices, t)

		msgs.SendSMS(s[0], m)
	}
}

func main() {
	hostname, err := os.Hostname()
	if err != nil {
		panic(fmt.Sprintf("Encountered error obtaining hostname: %s", err))
	}
	logger := zap.New(
		zap.NewJSONEncoder(),
		zap.Fields(zap.String("host", hostname)),
	)

	// Setup
	mainConfig, err := config.LoadFromFile(configFilepath)

	if err != nil {
		logger.Panic(
			"Encountered err loading config",
			zap.Error(err),
		)
	} else {
		logger.Info(
			"Successfully loaded config",
		)
	}

	srvs, err := services.LoadFromFile(servicesfilePath, logger)
	msgs := alerting.CreateMessaging(mainConfig.Messaging, srvs)

	if err != nil {
		logger.Panic(
			"Encountered err loading services",
			zap.Error(err),
		)
	} else {
		logger.Info(
			"Successfully loaded services",
		)
	}

	// welcomeOwners(srvs, msgs)

	// Start testing services
	testTicker := time.NewTicker(time.Second * time.Duration(mainConfig.TestInterval))

	go func(logger zap.Logger) {
		for _ = range testTicker.C {
			// run service tests
			for _, s := range srvs.Services {
				if err := s.RunAll(); err != nil {
					// Sanity check, do we have a Internet connection?
					if !hasInternet(logger) {
						a := alerting.CreateAlert("Peekaboo has no Internet connection.", logger)
						a.Log()
						continue
					}

					alert := alerting.CreateAlert(fmt.Sprintf("%s: %s. ID: %s", s.Name, err, s.ID), logger)

					for _, o := range s.Owners {
						if err := msgs.SendSMS(o.Cell, alert.Message); err != nil {
							a := alerting.CreateAlert(fmt.Sprintf("%s ID: %s", "Messaging failed to alert owners", s.ID), logger)
							a.Log()
						}
					}

					alert.Log()
				}
			}
		}
	}(logger)

	// run web service for handling notification replies
	msgs.RunReplyHandler()
}
