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

// assertArgPair verifies that flag is immediately followed by value in args.
func assertArgPair(t *testing.T, args []string, flag, value string) {
	t.Helper()
	for i, arg := range args {
		if arg == flag {
			require.Less(t, i+1, len(args), "flag %s has no value", flag)
			assert.Equal(t, value, args[i+1], "flag %s value mismatch", flag)
			return
		}
	}
	t.Errorf("flag %s not found in args", flag)
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

func TestSendAllFlags(t *testing.T) {
	bin, logFile := fakeBinary(t)

	p := &tn.TerminalNotifier{
		Path:         bin,
		Message:      "{{.Message}}",
		Title:        "{{.Title}}",
		Subtitle:     "sub",
		Sound:        "default",
		Group:        "grp",
		Open:         "https://example.com",
		Execute:      "echo hi",
		Activate:     "com.apple.Safari",
		Sender:       "com.apple.Terminal",
		AppIcon:      "/path/to/icon.png",
		ContentImage: "/path/to/image.png",
		IgnoreDnD:    true,
	}
	err := p.Send(context.Background(), notifier.Notification{
		Message: "hi",
		Title:   "test",
	})
	require.NoError(t, err)

	args := readArgs(t, logFile)
	assert.Contains(t, args, "-message")
	assert.Contains(t, args, "-title")
	assert.Contains(t, args, "-subtitle")
	assert.Contains(t, args, "sub")
	assert.Contains(t, args, "-sound")
	assert.Contains(t, args, "default")
	assert.Contains(t, args, "-group")
	assert.Contains(t, args, "grp")
	assert.Contains(t, args, "-open")
	assert.Contains(t, args, "https://example.com")
	assert.Contains(t, args, "-execute")
	assert.Contains(t, args, "echo hi")
	assert.Contains(t, args, "-activate")
	assert.Contains(t, args, "com.apple.Safari")
	assert.Contains(t, args, "-sender")
	assert.Contains(t, args, "com.apple.Terminal")
	assert.Contains(t, args, "-appIcon")
	assert.Contains(t, args, "/path/to/icon.png")
	assert.Contains(t, args, "-contentImage")
	assert.Contains(t, args, "/path/to/image.png")
	assert.Contains(t, args, "-ignoreDnD")

	// Verify flag-value pairing (flag immediately followed by its value)
	assertArgPair(t, args, "-message", "hi")
	assertArgPair(t, args, "-title", "test")
	assertArgPair(t, args, "-subtitle", "sub")
	assertArgPair(t, args, "-sound", "default")
}

func TestSendTemplateRendering(t *testing.T) {
	bin, logFile := fakeBinary(t)

	p := &tn.TerminalNotifier{
		Path:    bin,
		Message: "**{{.Project}}**: {{.Message}}",
		Title:   "{{.NotificationType}}: {{.Title}}",
	}
	err := p.Send(context.Background(), notifier.Notification{
		Message:          "Build complete",
		Title:            "Claude Code",
		Cwd:              "/home/user/myproject",
		NotificationType: "idle_prompt",
	})
	require.NoError(t, err)

	args := readArgs(t, logFile)
	assert.Contains(t, args, "**myproject**: Build complete")
	assert.Contains(t, args, "idle_prompt: Claude Code")
}

func TestSendWithVars(t *testing.T) {
	bin, logFile := fakeBinary(t)

	p := &tn.TerminalNotifier{
		Path:    bin,
		Message: "{{.Env}}: {{.Message}}",
		Title:   "{{.Title}}",
		Vars:    map[string]string{"env": "production"},
	}
	err := p.Send(context.Background(), notifier.Notification{
		Message: "done",
		Title:   "test",
	})
	require.NoError(t, err)

	args := readArgs(t, logFile)
	assert.Contains(t, args, "production: done")
}

func TestSendGroupDefaultsToSessionID(t *testing.T) {
	bin, logFile := fakeBinary(t)

	p := &tn.TerminalNotifier{}
	tn.ApplyDefaults(p)
	p.Path = bin
	err := p.Send(context.Background(), notifier.Notification{
		Message:   "hi",
		SessionID: "sess-42",
		Cwd:       "/tmp/proj",
	})
	require.NoError(t, err)

	args := readArgs(t, logFile)
	assert.Contains(t, args, "-group")
	assert.Contains(t, args, "sess-42")
}

func TestSendBadTemplate(t *testing.T) {
	bin, _ := fakeBinary(t)

	p := &tn.TerminalNotifier{
		Path:    bin,
		Message: "{{.Invalid",
	}
	err := p.Send(context.Background(), notifier.Notification{Message: "hi"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rendering message template")
}

func TestSendBinaryNotFound(t *testing.T) {
	p := &tn.TerminalNotifier{
		Path:    "/nonexistent/terminal-notifier",
		Message: "{{.Message}}",
	}
	err := p.Send(context.Background(), notifier.Notification{Message: "hi"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "running /nonexistent/terminal-notifier")
}

func TestSendBinaryNonZeroExit(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "terminal-notifier")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\necho 'error' >&2\nexit 1\n"), 0755))

	p := &tn.TerminalNotifier{
		Path:    script,
		Message: "{{.Message}}",
	}
	err := p.Send(context.Background(), notifier.Notification{Message: "hi"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "running")
}

func TestSendMinimalConfig(t *testing.T) {
	bin, logFile := fakeBinary(t)

	// Only path and message — no optional fields set
	p := &tn.TerminalNotifier{
		Path:    bin,
		Message: "{{.Message}}",
	}
	err := p.Send(context.Background(), notifier.Notification{Message: "hello"})
	require.NoError(t, err)

	args := readArgs(t, logFile)
	assert.Contains(t, args, "-message")
	assert.Contains(t, args, "hello")
	// No optional flags
	assert.NotContains(t, args, "-subtitle")
	assert.NotContains(t, args, "-sound")
	assert.NotContains(t, args, "-open")
	assert.NotContains(t, args, "-execute")
	assert.NotContains(t, args, "-activate")
	assert.NotContains(t, args, "-sender")
	assert.NotContains(t, args, "-appIcon")
	assert.NotContains(t, args, "-contentImage")
	assert.NotContains(t, args, "-ignoreDnD")
}
