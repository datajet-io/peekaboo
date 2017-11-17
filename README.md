# Peekaboo

Peekaboo is a extensible health check service for web endpoints/services.


## How it works

Peekaboo checks the health of a given web endpoint. The endpoint is healthy if...

- The API returns a specified GET request with HTTP status code 2xx
- The response is valid JSON (optional)


## Dependencies

* An account with one of the supported alerting providers; currently PagerDuty and Twilio
* Go >= 1.6

## Setup

At startup looks for a config file: `config.yml|yaml`.


### config.yml

Contains all the configuration for Peekaboo to run.

```yaml
alerters:
  pd-prod:
    type: pagerduty
    integration_key: SOMESECRETHERE

core:
  retry_timeout_seconds: 30 # optional
  test_interval_seconds: 60 # optional

services:
  some-endpoint:
    url: https://thisendpointdoesntexist.io/please/fail/me
    disabled: false
    tests:
      json: true
    handlers:
      - pd-prod
  another-endpoint:
    url: http://somefakeendpoint.io/fail/me
    disabled: false
    tests:
      json: true
    handlers:
      - pd-prod
```

## Building

Check out the Makefile.
