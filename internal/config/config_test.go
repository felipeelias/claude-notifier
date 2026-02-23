package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/felipeelias/claude-notifier/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	err := os.WriteFile(path, []byte(`
[global]
timeout = "5s"

[[notifiers.ntfy]]
url = "https://ntfy.example.com/topic1"

[[notifiers.ntfy]]
url = "https://ntfy.example.com/topic2"
`), 0644)
	require.NoError(t, err)

	cfg, err := config.Load(path)
	require.NoError(t, err)

	assert.Equal(t, 5*time.Second, cfg.Global.Timeout)
	require.Len(t, cfg.Notifiers, 1) // one key: "ntfy"
	require.Len(t, cfg.Notifiers["ntfy"], 2) // two instances
}

func TestLoadConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	err := os.WriteFile(path, []byte(`
[[notifiers.ntfy]]
url = "https://ntfy.example.com/topic"
`), 0644)
	require.NoError(t, err)

	cfg, err := config.Load(path)
	require.NoError(t, err)

	assert.Equal(t, 10*time.Second, cfg.Global.Timeout) // default
}

func TestLoadConfigMissing(t *testing.T) {
	_, err := config.Load("/nonexistent/config.toml")
	assert.Error(t, err)
}
