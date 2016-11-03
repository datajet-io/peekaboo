package alerting

import (
	"github.com/datajet-io/peekaboo/services"
)

type Alerter interface {
	TriggerAlert(service services.Service, alert Alert, details map[string]interface{}) error
}
