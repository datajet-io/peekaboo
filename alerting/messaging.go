package alerting

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/datajet-io/peekaboo/services"
)

//Twilio represents the configuration of the Twilio account used for in / out messaging
type Twilio struct {
	ReplyNumber string `json:"reply_number"`
	AccountSID  string `json:"account_sid"`
	AuthToken   string `json:"auth_token"`
	URLStr      string `json:"url_str"`
}

//MessagingConfig contains all configuration necessary for messaging
type MessagingConfig struct {
	Twilio               Twilio `json:"twilio"`
	ReplyHandlerPort     int    `json:"reply_handler_port"`
	ReplyHandlerCallback string `json:"reply_handler_callback"`
}

//Messaging handles send of alerts
type Messaging struct {
	config   MessagingConfig
	services *services.Services
}

//CreateMessaging set configuration for messaging
func CreateMessaging(m MessagingConfig, s *services.Services) *Messaging {

	return &Messaging{config: m, services: s}
}

//SendSMS sends the message to specificed cellnumber
func (m *Messaging) SendSMS(cellNumber string, message string) error {

	v := url.Values{}
	v.Set("To", cellNumber)
	v.Set("From", m.config.Twilio.ReplyNumber)

	v.Set("Body", message)
	rb := *strings.NewReader(v.Encode())

	// Create Client
	client := &http.Client{}

	req, _ := http.NewRequest("POST", m.config.Twilio.URLStr, &rb)
	req.SetBasicAuth(m.config.Twilio.AccountSID, m.config.Twilio.AuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Make request
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	if resp.StatusCode < 200 && resp.StatusCode > 300 {
		return errors.New("Failed to send messsage.")
	}

	return nil
}

//replyHandler handles owner's text message response to notifications
func (m *Messaging) replyHandler(w http.ResponseWriter, r *http.Request) {

	// deserialize URL request from Twilio

	v := r.URL.Query()

	body, ok := v["Body"]

	if !ok {
		// "Invalid request format by SMS relay. Body field missing."
		return
	}

	s, ok := v["From"]

	if !ok {
		// "Invalid request format by SMS relay. 'From' field missing."
		return
	}

	sender, err := m.services.GetOwnerByCell(s[0])

	if err != nil {
		return
	}

	// Parse reply

	query := strings.Split(body[0], " ")

	// check if valid format: command + service id
	if len(query) < 2 {
		m.SendSMS(sender.Cell, "Invalid response, usage: <command> <ID>. Valid commands: start, stop")
		return
	}

	command := query[0]

	srv, err := m.services.Get(query[1])

	if err != nil {
		m.SendSMS(sender.Cell, "No service found for code "+query[1])
		return
	}

	// process command

	switch command {
	case "start":
		{
			srv.Disabled = false
			m.SendSMS(sender.Cell, "Alerts for "+srv.Name+" started.")
			for _, o := range srv.Owners {
				if o.Cell != sender.Cell {
					m.SendSMS(o.Cell, "Alerts for "+srv.Name+" started by "+sender.Name)
				}
			}
		}
	case "stop":
		{
			srv.Disabled = true
			m.SendSMS(sender.Cell, "Alerts for "+srv.Name+" stopped.")
			for _, o := range srv.Owners {
				if o.Cell != sender.Cell {
					m.SendSMS(o.Cell, "Alerts for "+srv.Name+" stopped by "+sender.Name)
				}
			}

		}
	}

}

//RunReplyHandler starts accepting responses
func (m *Messaging) RunReplyHandler() {
	http.HandleFunc(m.config.ReplyHandlerCallback, m.replyHandler)

	port := strconv.Itoa(m.config.ReplyHandlerPort)

	http.ListenAndServe(":"+port, nil)
}
