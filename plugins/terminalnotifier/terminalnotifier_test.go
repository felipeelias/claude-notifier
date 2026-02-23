package terminalnotifier_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/felipeelias/claude-notifier/internal/notifier"
	tn "github.com/felipeelias/claude-notifier/plugins/terminalnotifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeBinary creates a shell script that logs all args to a file.
// Returns the script path and the log file path.
func fakeBinary(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "terminal-notifier")
	content := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$@\" > %s\n", logFile)
	require.NoError(t, os.WriteFile(script, []byte(content), 0755))
	return script, logFile
}

// readArgs reads the logged args from the fake binary.
func readArgs(t *testing.T, logFile string) []string {
	t.Helper()
	data, err := os.ReadFile(logFile)
	require.NoError(t, err)
	return strings.Split(strings.TrimSpace(string(data)), "\n")
}

func TestName(t *testing.T) {
	p := &tn.TerminalNotifier{}
	assert.Equal(t, "terminal-notifier", p.Name())
}

func TestDefaults(t *testing.T) {
	p := &tn.TerminalNotifier{}
	tn.ApplyDefaults(p)
	assert.Equal(t, "terminal-notifier", p.Path)
	assert.Equal(t, "{{.Message}}", p.Message)
	assert.Equal(t, "Claude Code ({{.Project}})", p.Title)
	assert.Equal(t, "{{.SessionID}}", p.Group)
}

func TestImplementsNotifier(t *testing.T) {
	var _ notifier.Notifier = &tn.TerminalNotifier{}
}

func TestSend(t *testing.T) {
	bin, logFile := fakeBinary(t)

	p := &tn.TerminalNotifier{
		Path:    bin,
		Message: "{{.Message}}",
		Title:   "{{.Title}}",
	}
	err := p.Send(context.Background(), notifier.Notification{
		Message: "Task complete",
		Title:   "Claude Code",
		Cwd:     "/home/user/myproject",
	})
	require.NoError(t, err)

	args := readArgs(t, logFile)
	assert.Contains(t, args, "-message")
	assert.Contains(t, args, "Task complete")
	assert.Contains(t, args, "-title")
	assert.Contains(t, args, "Claude Code")
}
