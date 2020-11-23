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

		assert.Equal(t, "test-user", c.HiveConfig.Username)
		assert.Equal(t, "test-password", c.HiveConfig.Password)
		assert.Equal(t, "id-000", c.HiveConfig.ThermostatID)

		assert.Equal(t, "influx-user", c.InfluxDBConfig.Username)
		assert.Equal(t, "influx-password", c.InfluxDBConfig.Password)
		assert.Equal(t, "my-database", c.InfluxDBConfig.Database)
	})
}
