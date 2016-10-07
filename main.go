package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/cenk/backoff"
	"github.com/datajet-io/peekaboo/alerting"
	"github.com/datajet-io/peekaboo/config"
	"github.com/datajet-io/peekaboo/services"
)

const configFilepath = "config.json" // path to config file
const servicesfilePath = "services.json"

//checks if there is Internet connection by pinging Google
func hasInternet() bool {
	operation := func() error {
		// TODO: turn this into a config item
		response, err := http.Get("https://www.google.com")

		if err != nil {
			return err
		}

		io.Copy(ioutil.Discard, response.Body)
		defer response.Body.Close()

		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return errors.New("Returned status code was not acceptable")
		}
		return err
	}

	err := backoff.Retry(operation, backoff.NewExponentialBackOff())
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
	// Setup
	mainConfig, err := config.LoadFromFile(configFilepath)

	if err != nil {
		fmt.Println(err)
		return
	}

	srvs, err := services.LoadFromFile(servicesfilePath)
	msgs := alerting.CreateMessaging(mainConfig.Messaging, srvs)

	if err != nil {
		fmt.Println("Configuration Error", err)
		return
	}

	// welcomeOwners(srvs, msgs)

	// Start testing services
	testTicker := time.NewTicker(time.Second * time.Duration(mainConfig.TestInterval))

	go func() {
		for _ = range testTicker.C {
			// run service tests
			for _, s := range srvs.Services {
				if err := s.RunAll(); err != nil {
					// Sanity check, do we have a Internet connection?
					if !hasInternet() {
						a := alerting.CreateAlert("Peekaboo has no Internet connection.")
						a.Log()
						continue
					}

					alert := alerting.CreateAlert(fmt.Sprintf("%s: %s. ID: %s", s.Name, err, s.ID))

					for _, o := range s.Owners {
						if err := msgs.SendSMS(o.Cell, alert.Message); err != nil {
							a := alerting.CreateAlert(fmt.Sprintf("%s ID: %s", "Messaging failed to alert owners", s.ID))
							a.Log()
						}
					}

					alert.Log()
				}
			}
		}
	}()

	// run web service for handling notification replies
	msgs.RunReplyHandler()
}
