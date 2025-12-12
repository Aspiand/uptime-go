package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"
	"uptime-go/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigParsing(t *testing.T) {
	configContent := `
monitor:
  - url: "https://example.com"
    enabled: true
    interval: "1m30s"
  - url: "https://google.com"
    enabled: true
    interval: "30s"
    response_time_threshold: "5s"
  - url: "https://guthib.com"
    interval: "30s"
    response_time_threshold: "5s"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "uptime.yml")

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg := config.New()
	err = cfg.Load(configPath)
	require.NoError(t, err)

	require.Len(t, cfg.Monitors, 2)

	assert.Equal(t, "https://example.com", cfg.Monitors[0].URL)
	assert.Equal(t, 90*time.Second, cfg.Monitors[0].Interval)

	assert.Equal(t, "https://google.com", cfg.Monitors[1].URL)
	assert.Equal(t, 30*time.Second, cfg.Monitors[1].Interval)
	assert.Equal(t, 5*time.Second, cfg.Monitors[1].ResponseTimeThreshold)
}
