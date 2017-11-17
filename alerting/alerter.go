package alerting

// NewAlerterClient is a helper function to create native alerter clients
func NewAlerterClient(name string, config map[string]interface{}) Alerter {
	var client Alerter

	switch config["type"] {
	case "pagerduty":
		client = NewPagerdutyClient(config["integration_key"].(string))
	default:
		panic("woops")
	}
	return client
}

// Alerter generic interface type
type Alerter interface {
	Trigger(serviceName string, alert *Alert) error
	Resolve(serviceName string, alert *Alert) error
}

//Alerters map of alerter handlers to use
type Alerters map[string]Alerter
