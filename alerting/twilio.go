package alerting

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/datajet-io/peekaboo/retry"
	"github.com/uber-go/zap"
)

//Twilio represents the configuration of the Twilio account used for in / out messaging
type TwilioClient struct {
	ReplyNumber          string            `json:"reply_number"`
	AccountSID           string            `json:"account_sid"`
	AuthToken            string            `json:"auth_token"`
	URLStr               string            `json:"url_str"`
	ReplyHandlerPort     int               `json:"reply_handler_port"`
	ReplyHandlerCallback string            `json:"reply_handler_callback"`
	Recipients           map[string]string `json:"recipients"`
}

//Twilio represents the configuration of the Twilio account used for in / out messaging
func NewTwilioClient(replyNumber, accountSID, authToken, urlStr, replyHandlerCallback string, replyHandlerPort int, recipients map[string]string) *TwilioClient {
	return &TwilioClient{
		ReplyNumber:          replyNumber,
		AccountSID:           accountSID,
		AuthToken:            authToken,
		URLStr:               urlStr,
		ReplyHandlerPort:     replyHandlerPort,
		ReplyHandlerCallback: replyHandlerCallback,
		Recipients:           recipients,
	}
}

func (t TwilioClient) TriggerAlert(serviceName string, logger zap.Logger, alert Alert, details map[string]interface{}) error {
	var err error

	for recipientKey, recipientValue := range t.Recipients {
		operation := func() error {
			v := url.Values{}
			v.Set("To", strings.Replace(recipientValue, " ", "", -1))
			v.Set("From", strings.Replace(t.ReplyNumber, " ", "", -1))

			v.Set("Body", alert.Message)
			rb := *strings.NewReader(v.Encode())

			// Create Client
			client := &http.Client{}

			// Create the request
			req, err := http.NewRequest("POST", t.URLStr, &rb)
			if err != nil {
				logger.Warn(
					"Encountered error creating request",
					zap.Error(err),
				)
				return err
			}

			req.SetBasicAuth(t.AccountSID, t.AuthToken)
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			// Make request
			resp, err := client.Do(req)

			if err != nil {
				logger.Warn(
					"Encountered error calling Twilio API",
					zap.Error(err),
				)
				return err
			}

			if resp.StatusCode < 200 && resp.StatusCode > 300 {
				logger.Warn(
					"Received unexpected response code",
					zap.Int("resp-code", resp.StatusCode),
					zap.String("recipient-name", recipientKey),
					zap.String("recipient-number", recipientValue),
					zap.Error(err),
				)
				return errors.New("Failed to send messsage.")
			}

			return err
		}

		err = retry.Retrying(operation, logger)
	}

	return err
}
