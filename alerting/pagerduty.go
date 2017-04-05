package alerting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/datajet-io/peekaboo/globals"
	"github.com/datajet-io/peekaboo/retry"
	"github.com/uber-go/zap"
)

const apiEndpoint = "https://events.pagerduty.com/generic/2010-04-15/create_event.json"

//PagerdutyClient represents the configuration of the Pageruty account used for notifications
type PagerdutyClient struct {
	integrationKey string
}

// PagerdutyEvent represents an API event object
type PagerdutyEvent struct {
	Service_key  string                 `json:"service_key"`
	Incident_key string                 `json:"incident_key"`
	Event_type   string                 `json:"event_type"`
	Description  string                 `json:"description"`
	Client       string                 `json:"client"`
	Details      map[string]interface{} `json:"details"`
}

// NewPagerdutyClient is an initilisation helper
func NewPagerdutyClient(key string) *PagerdutyClient {
	return &PagerdutyClient{integrationKey: key}
}

// Trigger is a helper function to wrap CreateEvent
func (p *PagerdutyClient) Trigger(serviceName string, alert *Alert, details map[string]interface{}) error {
	return p.CreateEvent(serviceName, alert, "trigger", details)
}

// Resolve is a helper function to wrap CreateEvent
func (p *PagerdutyClient) Resolve(serviceName string, alert *Alert, details map[string]interface{}) error {
	return p.CreateEvent(serviceName, alert, "resolve", details)
}

// CreateEvent is a generic func to create pagerduty events
func (p *PagerdutyClient) CreateEvent(serviceName string, alert *Alert, action string, details map[string]interface{}) error {
	eventData := &PagerdutyEvent{
		Service_key:  p.integrationKey,
		Incident_key: serviceName,
		Event_type:   action,
		Description:  alert.Message,
		Client:       "peekaboo",
		Details:      details,
	}

	eventDataJSON, err := json.Marshal(eventData)
	if err != nil {
		globals.Logger.Warn(
			"Error marshaling 'eventData' to JSON",
			zap.Object("eventData", eventData),
			zap.Error(err),
		)
		return err
	}

	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(eventDataJSON))
	if err != nil {
		globals.Logger.Warn(
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
			globals.Logger.Warn(
				"Encountered error calling Pagerduty API",
				zap.Error(err),
			)
			return err
		}
		defer resp.Body.Close()

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			globals.Logger.Warn(
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
			globals.Logger.Error(
				"Received 400 Invalid Event response",
				zap.String("resp-body", string(respBody)),
				zap.Error(err),
			)
			return nil
		case 403, 429:
			// This is retryable, so return an error
			return fmt.Errorf("Received '%d' rate limited response", statusCode)
		default:
			// This is might be retryable, unknown territory
			return fmt.Errorf("Received unexpected response, code was '%d', and body was: '%s'", statusCode, string(respBody))
		}
	}

	return retry.Retrying(operation, globals.Logger)
}
