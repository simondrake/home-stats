package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("should error with non-existent file", func(t *testing.T) {
		c, err := New("non-existent")

		assert.Nil(t, c)
		assert.EqualError(t, err, "unable to open file: open non-existent: no such file or directory")
	})

	t.Run("should not erorr with valid file", func(t *testing.T) {
		c, err := New("test_config.json")

		assert.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("should set correct values", func(t *testing.T) {
		a := assert.New(t)

		c, err := New("test_config.json")

		assert.NoError(t, err)
		assert.NotNil(t, c)

		// Thermostat config values
		a.True(c.Thermostat.Enabled)
		a.Equal("10m", c.Thermostat.Interval)
		a.Equal("user", c.Thermostat.Username)
		a.Equal("password", c.Thermostat.Password)
		a.Equal("000-111", c.Thermostat.ThermostatID)

		// Thermostat Boost config values
		a.False(c.Thermostat.AutoBoost.Enabled)
		a.Equal(18.0, c.Thermostat.AutoBoost.MinTemperature)
		a.Equal(int32(30), c.Thermostat.AutoBoost.TargetDuration)
		a.Equal(int32(24), c.Thermostat.AutoBoost.TargetTemperature)

		// Weather config values
		a.False(c.Weather.Enabled)
		a.Equal("3h", c.Weather.Interval)
		a.Equal("London", c.Weather.City)
		a.Equal("United Kingdom", c.Weather.Country)
		a.Equal("2222", c.Weather.APIKey)
		a.Equal("metric", c.Weather.Units)

		// Database config values
		a.Equal("http://localhost:3000", c.Database.URI)
		a.Equal("dbUser", c.Database.Username)
		a.Equal("dbPassword", c.Database.Password)
		a.Equal("db", c.Database.Database)
	})
}
