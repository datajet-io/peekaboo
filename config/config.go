package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/datajet-io/peekaboo/alerting"
)

// Config represents a Kanary configuration
type Config struct {
	Messaging    alerting.MessagingConfig `json:"messaging"`
	TestInterval int                      `json:"test_interval"`
}

//LoadFromFile loads the configuration from the given filepath
func LoadFromFile(filepath string) (*Config, error) {

	data, err := ioutil.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	cfg := &Config{}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, errors.New("Could not read " + filepath + ", unserialization failed.")
	}

	if cfg.TestInterval < 10 {
		return nil, errors.New("Test interval to short, must be 10 seconds or higher")
	}

	return cfg, nil
}
