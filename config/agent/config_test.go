package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("ADDRESS", "localhost:8080")
	t.Setenv("RATE_LIMIT", "2")
	t.Setenv("KEY", "blank/key")
	t.Setenv("REPORT_INTERVAL", "3")
	t.Setenv("POLL_INTERVAL", "4")
	t.Setenv("LOG_LEVEL", "DEBUG")

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotEmpty(t, cfg.FlagRunAddr)
	assert.NotEmpty(t, cfg.FlagLogLevel)
	assert.NotEmpty(t, cfg.FlagReportInterval)
	assert.Equal(t, 4, cfg.FlagPollInterval)
	assert.NotEmpty(t, cfg.FlagRateLimit)

	assert.Equal(t, cfg.FlagRunAddr, "localhost:8080")
	assert.Equal(t, cfg.FlagRateLimit, 2)
	assert.Equal(t, cfg.FlagHashKey, "blank/key")
	assert.Equal(t, cfg.FlagReportInterval, 3)
	assert.Equal(t, cfg.FlagPollInterval, 4)
	assert.Equal(t, cfg.FlagLogLevel, "DEBUG")
}
