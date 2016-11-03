package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
	alerterConfigs := make(map[string]map[string]alerting.Alerter)

	// Setup any alerters found in the config
	for alerterName, alerterValue := range mainConfig.Alerter {
		// Check if this alerter type is already present, otherwise create it
		if _, ok := alerterConfigs[alerterName]; !ok {
			alerterConfigs[alerterName] = make(map[string]alerting.Alerter)
		}

		switch alerterName {
		case "pagerduty":
			// TODO : extract this into the alerters initlisation function
			for alerterInstanceName, alerterInstanceValue := range alerterValue {
				alerterConfigs[alerterName][alerterInstanceName] = alerting.NewPagerdutyClient(alerterInstanceValue.(map[string]interface{})["integration-key"].(string))
			}
		case "twilio":
			// TODO : extract this into the alerters initlisation function
			for alerterInstanceName, alerterInstanceValue := range alerterValue {
				replyNumber := alerterInstanceValue.(map[string]interface{})["reply_number"].(string)
				accountId := alerterInstanceValue.(map[string]interface{})["account_sid"].(string)
				authToken := alerterInstanceValue.(map[string]interface{})["auth_token"].(string)
				urlStr := alerterInstanceValue.(map[string]interface{})["url_str"].(string)
				replyHandlerCallback := alerterInstanceValue.(map[string]interface{})["reply_handler_callback"].(string)
				replyHandlerPort := int(alerterInstanceValue.(map[string]interface{})["reply_handler_port"].(float64))

				recipients := make(map[string]string, 0)
				for recipientsKey, recipientsVal := range alerterInstanceValue.(map[string]interface{})["recipients"].(map[string]interface{}) {
					recipients[recipientsKey] = recipientsVal.(string)
				}

				alerterConfigs[alerterName][alerterInstanceName] = alerting.NewTwilioClient(
					replyNumber,
					accountId,
					authToken,
					urlStr,
					replyHandlerCallback,
					replyHandlerPort,
					recipients)
			}
		default:
			fmt.Printf("Provided alerter '%s' isn't supported, continuing.\n", alerterName)
			continue
		}
	}

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

	// Start testing services
	testTicker := time.NewTicker(time.Second * time.Duration(mainConfig.TestInterval))

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
				alert.Log()

				for _, a := range s.Alerters {
					switch alerter := a.Type; alerter {
					case "pagerduty":
						// TODO : refactor this into the alerter so it's more succinct
						if _, ok := alerterConfigs[alerter][a.Name]; ok {
							alerterConfigs[alerter][a.Name].TriggerAlert(*s, *alert, make(map[string]interface{}, 0))
						}
					case "twilio":
						// TODO : refactor this into the alerter so it's more succinct
						if _, ok := alerterConfigs[alerter][a.Name]; ok {
							alerterConfigs[alerter][a.Name].TriggerAlert(*s, *alert, make(map[string]interface{}, 0))
						}
					default:
						fmt.Printf("Provided alerter '%s' isn't supported, continuing.\n", alerter)
						continue
					}
				}
			}
		}
	}
}
