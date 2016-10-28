package alerting

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/datajet-io/peekaboo/retry"
	"github.com/datajet-io/peekaboo/services"
	"github.com/uber-go/zap"
)

const apiEndpoint = "https://events.pagerduty.com/generic/2010-04-15/create_event.json"

//Pagerduty represents the configuration of the Pageruty account used for notifications
type PagerdutyClient struct {
	integrationKey string
}

type PagerdutyEvent struct {
	Service_key  string                 `json:"service_key"`
	Incident_key string                 `json:"incident_key"`
	Event_type   string                 `json:"event_type"`
	Description  string                 `json:"description"`
	Client       string                 `json:"client"`
	Details      map[string]interface{} `json:"details"`
}

func NewPagerdutyClient(key string) *PagerdutyClient {
	return &PagerdutyClient{integrationKey: key}
}

func (p PagerdutyClient) TriggerAlert(service services.Service, alert Alert, details map[string]interface{}) error {
	eventData := &PagerdutyEvent{
		Service_key:  p.integrationKey,
		Incident_key: service.Name,
		Event_type:   "trigger",
		Description:  alert.Message,
		Client:       "peekaboo",
		Details:      details,
	}

	eventDataJson, err := json.Marshal(eventData)
	if err != nil {
		service.Logger.Warn(
			"Error marshaling 'eventData' to JSON",
			zap.Object("eventData", eventData),
			zap.Error(err),
		)
		return err
	}

	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(eventDataJson))
	if err != nil {
		service.Logger.Warn(
			"Encountered error creating request",
			zap.Error(err),
		)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Make request, with retrying support
	operation := func() error {
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			service.Logger.Warn(
				"Encountered error calling Pagerduty API",
				zap.Error(err),
			)
			return err
		}
		defer resp.Body.Close()

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			service.Logger.Warn(
				"Encountered error reading response body",
				zap.Error(err),
			)
			return err
		}

		statusCode := resp.StatusCode
		if statusCode == 200 {
			return nil
		}

		switch statusCode {
		case 400:
			service.Logger.Error(
				"Received 400 Invalid Event response",
				zap.String("resp-body", string(respBody)),
				zap.Error(err),
			)
			return nil
		case 403, 429:
			// This is retryable, so return an error
			return errors.New(fmt.Sprintf("Received '%s' rate limited response", statusCode))
		default:
			// This is might be retryable, unknown territory
			return errors.New(fmt.Sprintf("Received unexpected response, code was '%s', and body was: '%s'", statusCode, string(respBody)))
		}
		return err
	}

	return retry.Retrying(operation, service.Logger)
}
