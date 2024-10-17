package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("ADDRESS", "localhost:8080")
	t.Setenv("LOG_LEVEL", "INFO")
	t.Setenv("KEY", "blank/key")
	t.Setenv("STORE_INTERVAL", "3")
	t.Setenv("DATABASE_DSN", "temp/dsn")
	t.Setenv("RESTORE", "true")

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotEmpty(t, cfg.FlagRunAddr)
	assert.NotEmpty(t, cfg.FlagLogLevel)
	assert.NotEmpty(t, cfg.FlagStoreInterval)
	assert.NotEmpty(t, cfg.FlagStorePath)
	assert.NotEmpty(t, cfg.FlagLogLevel)

	assert.Equal(t, cfg.FlagRunAddr, "localhost:8080")
	assert.Equal(t, cfg.FlagLogLevel, "INFO")
	assert.Equal(t, cfg.FlagHashKey, "blank/key")
	assert.Equal(t, cfg.FlagStoreInterval, 3)
	assert.Equal(t, cfg.FlagDBDSN, "temp/dsn")
	assert.Equal(t, cfg.FlagRestore, true)
}
