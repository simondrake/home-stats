package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	Thermostat ThermostatConfig `json:"thermostat,omitempty"`
	Weather    WeatherConfig    `json:"weather,omitempty"`
	Database   DatabaseConfig   `json:"database,omitempty"`
}

type ThermostatConfig struct {
	Enabled      bool      `json:"enabled,omitempty"`
	Interval     string    `json:"interval,omitempty"`
	Username     string    `json:"username,omitempty"`
	Password     string    `json:"password,omitempty"`
	ThermostatID string    `json:"thermostatID,omitempty"`
	AutoBoost    AutoBoost `json:"autoBoost,omitempty"`
	HiveSSO      HiveSSO   `json:"hiveSSO,omitempty"`
}

type AutoBoost struct {
	Enabled           bool    `json:"enabled,omitempty"`
	MinTemperature    float64 `json:"minTemperature,omitempty"`
	TargetDuration    int32   `json:"targetDuration,omitempty"`
	TargetTemperature int32   `json:"targetTemperature,omitempty"`
}

type HiveSSO struct {
	PoolID                string `json:"poolID,omitempty"`
	PublicCognitoClientID string `json:"publicCognitoClientID,omitempty"`
}

type WeatherConfig struct {
	Enabled  bool   `json:"enabled,omitempty"`
	Interval string `json:"interval,omitempty"`
	City     string `json:"city,omitempty"`
	Country  string `json:"country,omitempty"`
	APIKey   string `json:"apiKey,omitempty"`
	Units    string `json:"units,omitempty"`
}

type SpeedTestConfig struct {
	Enabled  bool   `json:"enabled,omitempty"`
	Interval string `json:"interval,omitempty"`
}

type DatabaseConfig struct {
	URI      string `json:"uri,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Database string `json:"database,omitempty"`
}

func New(fileName string) (*Config, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}

	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}

	c := &Config{}
	if err := json.Unmarshal(b, c); err != nil {
		return nil, fmt.Errorf("unable to unmarshal to Config: %w", err)
	}

	return c, nil
}
