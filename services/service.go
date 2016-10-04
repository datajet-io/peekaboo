package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/datajet-io/peekaboo/retry"
)

//Owner represents the owner of a service who will be contacted in case of test failure
type Owner struct {
	Name string `json:"name"`
	Cell string `json:"cell"`
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
	ID     string  // unique random ID assinged at run-time
	Name   string  `json:"name"`
	URL    string  `json:"url"`
	Tests  Test    `json:"tests"`
	Owners []Owner `json:"owners"`
	Active bool
}

//RunAll performs all tests for the given service
func (s *Service) RunAll() error {

	if !s.Active {
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

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("Service HTTP status code is %d, %d expected", response.StatusCode, http.StatusOK)
		}
		return err
	}

	return retry.Retrying(operation)
	// return backoff.Retry(operation, &defaultBackoff)
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
			return errors.New("Could not unserialize response into JSON")
		}

		return nil
	}

	return retry.Retrying(operation)
	// return backoff.Retry(operation, &defaultBackoff)
}

//Time tests if the service responds within the specified time limit
func (s *Service) Time() error {
	operation := func() error {
		start := time.Now()

		response, err := http.Get(s.URL)

		if err != nil {
			return errors.New("Could not reach API")
		}

		defer response.Body.Close()

		elapsedMilliseconds := time.Since(start).Nanoseconds() / 1000000

		if elapsedMilliseconds > int64(s.Tests.MaxResponseTime) {
			return fmt.Errorf("Service response time is too damm high. Current %dms, <%dms expected", elapsedMilliseconds, s.Tests.MaxResponseTime)
		}

		return nil
	}

	return retry.Retrying(operation)
	// return backoff.Retry(operation, &defaultBackoff)
}
