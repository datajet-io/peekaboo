package alerting

import (
	"fmt"
	"time"

	"github.com/datajet-io/peekaboo/globals"
	"go.uber.org/zap"
)

//Alert represents an Alert message
type Alert struct {
	CreatedAt time.Time
	Message   string
}

//NewAlert creates an alert
func NewAlert(message string) *Alert {
	alert := Alert{
		CreatedAt: time.Now(),
		Message:   fmt.Sprintf("%s", message),
	}

	globals.Logger.Warn(
		alert.Message,
		zap.Time("created-at", alert.CreatedAt),
	)

	return &alert
}
