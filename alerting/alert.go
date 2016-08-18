package alerting

import (
	"fmt"
	"time"
)

//Alert represetns an Alert message
type Alert struct {
	CreatedAt time.Time
	Message   string
}

//CreateAlert creates an alert
func CreateAlert(message string) *Alert {
	return &Alert{
		CreatedAt: time.Now(),
		Message:   fmt.Sprintf("%s", message),
	}
}

//Log outputs an Alert to stdout with timestamp
func (a *Alert) Log() {
	time := fmt.Sprintf("%d/%d/%d %02d:%02d:%02d", a.CreatedAt.Month(), a.CreatedAt.Day(), a.CreatedAt.Year(), a.CreatedAt.Hour(), a.CreatedAt.Minute(), a.CreatedAt.Second())
	fmt.Printf("%s\t%s\n", time, a.Message)
}
