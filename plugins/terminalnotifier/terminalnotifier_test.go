package terminalnotifier_test

import (
	"testing"

	"github.com/felipeelias/claude-notifier/internal/notifier"
	tn "github.com/felipeelias/claude-notifier/plugins/terminalnotifier"
	"github.com/stretchr/testify/assert"
)

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
