package main_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testBinary string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "claude-notifier-test-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	testBinary = filepath.Join(tmp, "claude-notifier")
	build := exec.Command("go", "build", "-o", testBinary, ".")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		log.Fatalf("failed to build binary: %v", err)
	}

	os.Exit(m.Run())
}

func TestEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Start mock ntfy server
	var gotBody string
	var gotTitle string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		gotTitle = r.Header.Get("Title")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Write config pointing to mock server
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.toml")
	configContent := `[[notifiers.ntfy]]
url = "` + srv.URL + `"
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Prepare stdin JSON
	input, err := json.Marshal(map[string]string{
		"message": "Build complete",
		"title":   "Claude Code (myproject)",
		"cwd":     "/home/user/myproject",
	})
	require.NoError(t, err)

	// Run claude-notifier
	cmd := exec.Command(testBinary, "--config", configPath)
	cmd.Stdin = bytes.NewReader(input)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	require.NoError(t, err, "stderr: %s", stderr.String())

	// Verify ntfy received the notification
	assert.Equal(t, "Build complete", gotBody)
	assert.Equal(t, "Claude Code (myproject)", gotTitle)
}

func TestEndToEndMissingConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Run with nonexistent config — should still exit 0 (never fail the hook)
	input, _ := json.Marshal(map[string]string{"message": "test"})
	cmd := exec.Command(testBinary, "--config", "/nonexistent/config.toml")
	cmd.Stdin = bytes.NewReader(input)
	err := cmd.Run()
	assert.NoError(t, err, "should exit 0 even with missing config")
}

func TestEndToEndInitCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	configPath := filepath.Join(t.TempDir(), "claude-notifier", "config.toml")
	cmd := exec.Command(testBinary, "--config", configPath, "init")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	require.NoError(t, err)

	assert.Contains(t, stdout.String(), "Config created")

	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "[[notifiers.ntfy]]")
}

func TestEndToEndTestCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Start mock server
	var received bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	configPath := filepath.Join(t.TempDir(), "config.toml")
	os.WriteFile(configPath, []byte(`[[notifiers.ntfy]]
url = "`+srv.URL+`"
`), 0644)

	cmd := exec.Command(testBinary, "--config", configPath, "test")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	require.NoError(t, err)

	assert.True(t, received, "mock server should have received the test notification")
	assert.Contains(t, stdout.String(), "Test notification sent successfully")
}
