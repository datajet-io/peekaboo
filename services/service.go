package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cenk/backoff"
	"github.com/datajet-io/peekaboo/alerting"
	"github.com/datajet-io/peekaboo/globals"
	"github.com/datajet-io/peekaboo/retry"
	"github.com/uber-go/zap"
)

// InternetCheckURL is the URL to use to check internet connectivity
const InternetCheckURL string = "https://www.google.com"

//Test contains all the tests parameters for a service
type Test struct {
	RetryTimeoutSeconds int
	ValidateJSON        bool `json:"json"`
}

//Service represents the API to test
type Service struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Disabled bool   `json:"disabled"`
	Failed   bool
	Tests    Test               `json:"tests"`
	Handlers []alerting.Alerter `json:"alerters"`
}

//checks if there is Internet connection by pinging Google
func hasInternet() bool {
	notify := func(err error, duration time.Duration) {
		globals.Logger.Warn(
			"Encountered error checking for internet connection",
			zap.Int64("duration", duration.Nanoseconds()/int64(time.Millisecond)),
			zap.Error(err),
		)
	}

	operation := func() error {
		response, err := http.Get(InternetCheckURL)

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

//PerformChecks calls RunAll and handles calling the alerters as needed
func (s *Service) PerformChecks() {
	if err := s.RunAll(); err != nil {
		// Sanity check, do we have an Internet connection?
		if !hasInternet() {
			alerting.NewAlert("Peekaboo has no Internet connection.")
			return
		}

		alert := alerting.NewAlert(fmt.Sprintf("%s: %s.", s.Name, err))

		for _, a := range s.Handlers {
			a.Trigger(s.Name, alert, make(map[string]interface{}, 0))
		}

		s.Failed = true
		return
	}

	// Failed is initialised to True, since we'd like to resolve any open
	//  alerts, if any, and if this is the first check of this service.
	if s.Failed {
		s.Failed = false

		alert := alerting.NewAlert(fmt.Sprintf("%s: is healthy.", s.Name))

		for _, a := range s.Handlers {
			a.Resolve(s.Name, alert, make(map[string]interface{}, 0))
		}

	}
	return
}

//RunAll performs all tests for the given service
func (s *Service) RunAll() error {
	if s.Disabled {
		return nil
	}

	var elapsedMilliseconds int64
	var err error
	var data []byte
	var response *http.Response

	operation := func() error {
		start := time.Now()
		response, err = http.Get(s.URL)

		if err != nil {
			globals.Logger.Warn(
				"Encountered error while retrieving URL",
				zap.String("url", s.URL),
				zap.Error(err),
			)
			return err
		}

		if err := s.Ping(response); err != nil {
			return err
		}

		data, err = ioutil.ReadAll(response.Body)
		elapsedMilliseconds = time.Since(start).Nanoseconds() / int64(time.Millisecond)

		return nil
	}

	if err = retry.Retrying(operation, s.Tests.RetryTimeoutSeconds, globals.Logger); err != nil {
		return err
	}

	if s.Tests.ValidateJSON {
		if err := s.JSON(data); err != nil {
			return err
		}
	}

	defer response.Body.Close()

	return nil
}

//Ping checks if a service is available
func (s *Service) Ping(response *http.Response) error {
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status was %d, but expected %d", response.StatusCode, http.StatusOK)
	}

	return nil
}

//JSON test if the service response is valid json
func (s *Service) JSON(data []byte) error {
	var d interface{}

	if err := json.Unmarshal(data, &d); err != nil {
		globals.Logger.Warn(
			"Could not unserialize response into JSON",
			zap.Error(err),
		)
		return err
	}

	return nil
}
