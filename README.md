# Peekaboo

Peekaboo is a text message-based health check service for restful APIs.

###How it works

Peekaboo checks the health of a given API endpoint. The endpoint is healthy if...

- The API returns a specified GET request with HTTP status code 2xx
- The response size is equal or exceeds the specified minimum size in kb
- The response time (ms) doesn't exceed the specified upper boundary
- The response is valid JSON (optional)

If any of the above fails, Peekaboo sends a text-message to the specified recipients unless the error is due to local network failure. Peekaboo will continue to issue alerts unless it is silenced by responding to the alert text message with ```stop <ID>``` with the ID being the alert id provided in the alert message. ```start <ID>``` (re)starts the alert.

####Dependencies

* Peekaboo uses Twillio for sending text-messages. Twillio credentials must be provided as enviroment variables.
* Go >1.6

####Setup

Peekabook runs as binary executable, at startup it looks for two files: ```services.json```and ```config.json```. Both files must be placed in the same folder as the Peekaboo binary.

##### services.json

Contains all the services to test and their respective owners 

```json
[
    {
        "name": "My Great Service",
        "url": "https://mygreatservice.com/api",
        "tests": {
            "min_response_size": 23,
            "max_response_time": 200,
            "json": true,
            "cert": false
        },
        "owners": [
            {
                "name": "Petar",
                "cell": "+491721727438"
            }
        ]
    }
]
```

Possible values:
* ```min_response_size```: Minimum size (kb) of the response from the GET request
* ```max_response_time```: Maximum time in ms Peekaboo will wait for a response
* ```json```: Test if returned response is valid JSON
* ```cert```: Test if the service's SSL certificate is valid and doesn't expire in the next 30 days
* ```owners```: People that will receive a text message if any of the tests for a given service fails. 

##### config.json

Contains configuration settings related to messaging / SMS relay

```json
{
    "messaging": {
        "twilio": {
            "reply_number": "+4915735983506",
            "account_sid": "AC062ca3c824e72bfab0d135b9553947d8",
            "auth_token": "9b7c2ff4c4b5aa9aefb3fe600f0a126f",
            "url_str": "https://api.twilio.com/2010-04-01/Accounts/AC062ca3c824e72bfab0d135b9553947d8/Messages.json"
        },
        "reply_handler_port": 8080,
        "reply_handler_callback": "/handler/replies"
    },
    "test_interval": 60
}
```

Possible values:
* ```twilio```: Your Twilio credentials
* ```reply_handler_port```: Peekaboo runs a web service to retrieve Twilio callbacks when owners respond to an alert. In production this needs to be on port 80 to be callable by Twilio.
* ```reply_handler_callback```: Path to the callback handler
* ```test_interval```: Time interval to run tests  (seconds)

###Todo

* Test for CERT expiration
* Multiple retries after service fails