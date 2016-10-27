package alerting

import (
	"fmt"
	"time"

	"github.com/uber-go/zap"
)

//Alert represetns an Alert message
type Alert struct {
	CreatedAt time.Time
	Logger    zap.Logger
	Message   string
}

//CreateAlert creates an alert
func CreateAlert(message string, logger zap.Logger) *Alert {
	return &Alert{
		CreatedAt: time.Now(),
		Logger:    logger,
		Message:   fmt.Sprintf("%s", message),
	}
}

//Log outputs an Alert to stdout with timestamp
func (a *Alert) Log() {
	a.Logger.Warn(
		a.Message,
		zap.Time("created-at", a.CreatedAt),
	)
}
