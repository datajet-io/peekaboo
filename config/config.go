package config

import (
	"fmt"
	"os"

	"github.com/datajet-io/peekaboo/alerting"
	"github.com/datajet-io/peekaboo/globals"
	"github.com/datajet-io/peekaboo/services"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// Config represents the configuration
type Config struct {
	Alerters alerting.Alerters
	Core     *Core
	Services services.Services
}

// Core is the type for our core configuration section
type Core struct {
	RetryTimeoutSeconds int
	TestIntervalSeconds int
}

func NewCore() *Core {
	// Set defaults
	defaultRetryTimeoutSeconds := 30
	defaultTestIntervalSeconds := 60

	return &Core{
		RetryTimeoutSeconds: defaultRetryTimeoutSeconds,
		TestIntervalSeconds: defaultTestIntervalSeconds,
	}
}

// Setup is responsible for initialising the configuration
func Setup() *Config {
	hostname, hostnameErr := os.Hostname()
	if hostnameErr != nil {
		panic(fmt.Sprintf("Encountered error obtaining hostname: %s", hostnameErr))
	}
	globals.Logger = zap.New(
		zap.NewJSONEncoder(),
		zap.Fields(zap.String("host", hostname)),
	)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viperErr := viper.ReadInConfig()
	if viperErr != nil {
		panic(fmt.Errorf("fatal error config file: %s", viperErr))
	}

	if !viper.IsSet("alerters") {
		globals.Logger.Panic("Error loading config, alerters section not present")
	} else if !viper.IsSet("services") {
		globals.Logger.Panic("Error loading config, services section not present")
	}
	tempConfig := viper.AllSettings()

	var config Config
	config.Core = SetupCore(tempConfig)
	config.Alerters = SetupAlerters(config, tempConfig)
	config.Services = SetupServices(config, tempConfig, config.Alerters)

	return &config
}

// SetupAlerters is responsible for parsing the alerters config section
func SetupAlerters(config Config, rawConfig map[string]interface{}) alerting.Alerters {
	alerters := make(alerting.Alerters, 0)

	for k, v := range rawConfig["alerters"].(map[string]interface{}) {
		alerters[k] = alerting.NewAlerterClient(k, v.(map[string]interface{}))
	}

	return alerters
}

// SetupCore is responsible for parsing the core config section
func SetupCore(config map[string]interface{}) *Core {
	core := NewCore()

	for k, v := range config["core"].(map[string]interface{}) {
		switch k {
		case "test_interval_seconds":
			if v, ok := v.(int); ok && v >= 1 {
				core.TestIntervalSeconds = v
			}
		case "retry_timeout_seconds":
			if v, ok := v.(int); ok && v >= 1 {
				core.RetryTimeoutSeconds = v
			}
		}
	}

	return core
}

// SetupServices is responsible for parsing the services config section
func SetupServices(config Config, rawConfig map[string]interface{}, alerters alerting.Alerters) services.Services {
	servicesObjects := make(map[string]*services.Service, 0)

	for k, v := range rawConfig["services"].(map[string]interface{}) {
		var service services.Service
		service.Name = k
		service.Failed = true

		for attributeKey, attributeValue := range v.(map[string]interface{}) {
			switch attributeKey {
			case "disabled":
				service.Disabled = attributeValue.(bool)
			case "handlers":
				service.Handlers = ParseHandlers(attributeValue.([]interface{}), alerters)
			case "tests":
				service.Tests = ParseTests(config, attributeValue.(map[string]interface{}))
			case "url":
				service.URL = attributeValue.(string)
			}
		}

		servicesObjects[k] = &service
	}

	return servicesObjects
}

// ParseHandlers is a helper function used by Setup
func ParseHandlers(handlers []interface{}, alerters alerting.Alerters) []alerting.Alerter {
	handlersObjects := make([]alerting.Alerter, 0)

	for _, v := range handlers {
		if _, ok := alerters[v.(string)]; !ok {
			globals.Logger.Error("referenced alerter doesn't exist",
				zap.String("alerter", v.(string)),
			)
			continue
		}
		handlersObjects = append(handlersObjects, alerters[v.(string)])
	}

	return handlersObjects
}

// ParseTests is a helper function used by Setup
func ParseTests(config Config, tests map[string]interface{}) services.Test {
	var testsObject services.Test
	testsObject.RetryTimeoutSeconds = config.Core.RetryTimeoutSeconds

	for k, v := range tests {
		switch k {
		case "cert":
			testsObject.ValidateCERT = v.(bool)
		case "json":
			testsObject.ValidateJSON = v.(bool)
		case "max_response_time":
			testsObject.MaxResponseTime = v.(int)
		case "min_response_size":
			testsObject.MinPayloadSize = v.(int)
		}
	}

	return testsObject
}
