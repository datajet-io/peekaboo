package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

const minTestInterval = 1

// Config represents a Kanary configuration
type Config struct {
	Alerter      map[string]map[string]interface{} `json:"alerter"`
	TestInterval int                               `json:"test_interval"`
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

	if cfg.TestInterval < minTestInterval {
		return nil, errors.New(fmt.Sprintf("Test interval to short, must be %s second(s) or higher", minTestInterval))
	}

	return cfg, nil
}
