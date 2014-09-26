package config

import (
	log "github.com/cihub/seelog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfigShouldNotParseInvalidYAML(t *testing.T) {
	assert := assert.New(t)
	var data = `
---
data_engine:
  mongodb:
	use_sharding: false
`
	config, err := Load([]byte(data))
	assert.NotNil(err)
	assert.Nil(config)
}

func TestConfigShouldParseDataEngineInfluxDB(t *testing.T) {
	assert := assert.New(t)

	var data = `
---
data_engine:
  type: influxdb
  influxdb:
    host: localhost
    port: 4242
    user: influx
    password: secret
    database: morgoth
`

	config, err := Load([]byte(data))
	require.Nil(t, err, "Error loading config: %v\n", err)

	log.Debugf("Config: %v", config)
	assert.Equal(config.DataEngine.InfluxDB.Host, "localhost")
	assert.Equal(config.DataEngine.InfluxDB.Port, 4242)
	assert.Equal(config.DataEngine.InfluxDB.User, "influx")
	assert.Equal(config.DataEngine.InfluxDB.Password, "secret")
	assert.Equal(config.DataEngine.InfluxDB.Database, "morgoth")
	assert.Equal(config.DataEngine.Type, InfluxDB)

	assert.Nil(config.Validate())
}

func TestConfigShouldParseDataEngineMongoDB(t *testing.T) {
	assert := assert.New(t)

	var data = `
---
data_engine:
  type: mongodb
  mongodb:
    host: localhost
    port: 27017
    database: morgoth
    use_sharding: false
`

	config, err := Load([]byte(data))
	require.Nil(t, err, "Error loading config: %v\n", err)

	log.Debugf("Config: %v", config)
	assert.Equal(config.DataEngine.MongoDB.Host, "localhost")
	assert.Equal(config.DataEngine.MongoDB.Port, 27017)
	assert.Equal(config.DataEngine.MongoDB.Database, "morgoth")
	assert.Equal(config.DataEngine.MongoDB.IsSharded, false)
	assert.Equal(config.DataEngine.Type, MongoDB)

	assert.Nil(config.Validate())
}

func TestConfigShouldNotValidateBadDataEngineConf(t *testing.T) {
	assert := assert.New(t)

	var data = `
---
data_engine:
  bad_key: 1
`
	_, err := Load([]byte(data))
	assert.NotNil(err)

}
