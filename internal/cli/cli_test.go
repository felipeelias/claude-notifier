package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	appcli "github.com/felipeelias/claude-notifier/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCommand(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "claude-notifier", "config.toml")

	app := appcli.New("test")
	err := app.Run([]string{"claude-notifier", "--config", configPath, "init"})
	require.NoError(t, err)

	_, err = os.Stat(configPath)
	assert.NoError(t, err, "config file should exist")

	content, _ := os.ReadFile(configPath)
	assert.Contains(t, string(content), "[global]")
}

func TestInitCommandAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	os.WriteFile(configPath, []byte("existing"), 0644)

	app := appcli.New("test")
	err := app.Run([]string{"claude-notifier", "--config", configPath, "init"})
	assert.Error(t, err, "should fail if config already exists")
}

func TestVersionFlag(t *testing.T) {
	var buf bytes.Buffer
	app := appcli.New("1.2.3")
	app.Writer = &buf

	err := app.Run([]string{"claude-notifier", "--version"})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "1.2.3")
}
