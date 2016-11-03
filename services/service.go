package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/datajet-io/peekaboo/retry"
	"github.com/uber-go/zap"
)

//Handler contains which notifiers to use for this service if it alerts
type Alerter struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

//Test contains all the tests parameters for a service
type Test struct {
	MaxResponseTime int  `json:"max_response_time"` // milliseconds
	MinPayloadSize  int  `json:"min_response_size"` // kilobyte
	ValidateJSON    bool `json:"json"`
	ValidateCERT    bool `json:"cert"`
}

//Service represents the API to test
type Service struct {
	ID       string    // unique random ID assinged at run-time
	Name     string    `json:"name"`
	URL      string    `json:"url"`
	Disabled bool      `json:"disabled"`
	Tests    Test      `json:"tests"`
	Alerters []Alerter `json:"alerters"`
	Logger   zap.Logger
}

//RunAll performs all tests for the given service
func (s *Service) RunAll() error {
	if s.Disabled {
		return nil
	}

	if err := s.Ping(); err != nil {
		return err
	}

	if err := s.JSON(); err != nil {
		return err
	}

	if err := s.Time(); err != nil {
		return err
	}

	return nil
}

//Ping checks if a service is available
func (s *Service) Ping() error {

	// Make request
	operation := func() error {
		response, err := http.Get(s.URL)

		if err != nil {
			return err
		}

		defer response.Body.Close()
		io.Copy(ioutil.Discard, response.Body)

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("Service HTTP status code is %d, %d expected", response.StatusCode, http.StatusOK)
		}
		return err
	}

	return retry.Retrying(operation, s.Logger)
}

//JSON test if the service response is valid json
func (s *Service) JSON() error {

	operation := func() error {
		response, err := http.Get(s.URL)
		if err != nil {
			return errors.New("Service not reachable")
		}

		defer response.Body.Close()
		data, err := ioutil.ReadAll(response.Body)

		if err != nil {
			return err
		}

		var d interface{}

		if err := json.Unmarshal(data, &d); err != nil {
			s.Logger.Warn(
				"Could not unserialize response into JSON",
				zap.Error(err),
			)
			return err
		}

		return nil
	}

	return retry.Retrying(operation, s.Logger)
}

//Time tests if the service responds within the specified time limit
func (s *Service) Time() error {
	operation := func() error {
		start := time.Now()

		response, err := http.Get(s.URL)

		if err != nil {
			s.Logger.Warn(
				"Encountered error while retrieving URL",
				zap.String("url", s.URL),
				zap.Error(err),
			)
			return err
		}

		defer response.Body.Close()
		io.Copy(ioutil.Discard, response.Body)

		elapsedMilliseconds := time.Since(start).Nanoseconds() / int64(time.Millisecond)

		if elapsedMilliseconds > int64(s.Tests.MaxResponseTime) {
			s.Logger.Warn(
				"Response time is too high",
				zap.String("url", s.URL),
				zap.Int64("resp_time", elapsedMilliseconds),
				zap.Int64("resp_limit", int64(s.Tests.MaxResponseTime)),
			)
			return fmt.Errorf("Service response time is too damm high. Current %dms, <%dms expected", elapsedMilliseconds, s.Tests.MaxResponseTime)
		}

		return nil
	}

	return retry.Retrying(operation, s.Logger)
}
