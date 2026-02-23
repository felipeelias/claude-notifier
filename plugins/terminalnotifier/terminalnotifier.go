package terminalnotifier

import (
	"context"
	"fmt"

	"github.com/felipeelias/claude-notifier/internal/cli"
	"github.com/felipeelias/claude-notifier/internal/notifier"
)

// TerminalNotifier sends macOS desktop notifications via the terminal-notifier binary.
type TerminalNotifier struct {
	Path         string            `toml:"path"`
	Message      string            `toml:"message"`
	Title        string            `toml:"title"`
	Subtitle     string            `toml:"subtitle"`
	Sound        string            `toml:"sound"`
	Group        string            `toml:"group"`
	Open         string            `toml:"open"`
	Execute      string            `toml:"execute"`
	Activate     string            `toml:"activate"`
	Sender       string            `toml:"sender"`
	AppIcon      string            `toml:"app_icon"`
	ContentImage string            `toml:"content_image"`
	IgnoreDnD    bool              `toml:"ignore_dnd"`
	Vars         map[string]string `toml:"vars"`
}

// ApplyDefaults sets sane defaults on a new TerminalNotifier instance.
func ApplyDefaults(n *TerminalNotifier) {
	n.Path = "terminal-notifier"
	n.Message = "{{.Message}}"
	n.Title = "Claude Code ({{.Project}})"
	n.Group = "{{.SessionID}}"
}

func (n *TerminalNotifier) Name() string { return "terminal-notifier" }

func (n *TerminalNotifier) Send(ctx context.Context, notif notifier.Notification) error {
	// Stub — implemented in Task 3
	return fmt.Errorf("not implemented")
}

func init() {
	if err := cli.Registry.Register("terminal-notifier", func() notifier.Notifier {
		n := &TerminalNotifier{}
		ApplyDefaults(n)
		return n
	}); err != nil {
		panic(err)
	}
}
