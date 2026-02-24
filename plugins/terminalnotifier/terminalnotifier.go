package terminalnotifier

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/felipeelias/claude-notifier/internal/notifier"
	"github.com/felipeelias/claude-notifier/internal/tmpl"
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

// SampleConfig returns example TOML configuration.
func (n *TerminalNotifier) SampleConfig() string {
	return `## macOS desktop notifications via terminal-notifier
## https://github.com/julienXX/terminal-notifier
[[notifiers.terminal-notifier]]

## Path to the terminal-notifier binary (required)
## Install: brew install terminal-notifier
path = "terminal-notifier"

## Go template for the message body
## Available variables: {{.Message}}, {{.Title}}, {{.Project}}, {{.Cwd}},
## {{.NotificationType}}, {{.SessionID}}, {{.TranscriptPath}}
## Custom variables from [notifiers.terminal-notifier.vars] are also available, title-cased
# message = "{{.Message}}"

## Go template for the notification title
# title = "Claude Code ({{.Project}})"

## Go template for the notification subtitle
# subtitle = ""

## Sound to play ("default" for system default, or a sound name from /System/Library/Sounds)
# sound = ""

## Group ID — only one notification per group is shown, replacing previous ones
## Defaults to session ID so notifications from the same session replace each other
# group = "{{.SessionID}}"

## URL or custom scheme to open when clicking the notification
## Not templated — static string for security (terminal-notifier executes these)
# open = ""

## Shell command to run when clicking the notification
## Not templated — static string for security (terminal-notifier executes these)
# execute = ""

## App bundle ID to activate when clicking the notification
# activate = ""

## Fake sender app bundle ID
# sender = ""

## Path to a custom app icon image
# app_icon = ""

## Path to an image to attach inside the notification
# content_image = ""

## Send notification even when Do Not Disturb is active
# ignore_dnd = false

## User-defined template variables
## Keys are title-cased for template access (env -> {{.Env}})
# [notifiers.terminal-notifier.vars]
# env = "production"
`
}

func (n *TerminalNotifier) buildArgs(tctx map[string]string) ([]string, error) {
	type tmplField struct {
		name, value, fallback, flag string
		required                    bool
	}
	fields := []tmplField{
		{"message", n.Message, "{{.Message}}", "-message", true},
		{"title", n.Title, "Claude Code ({{.Project}})", "-title", false},
		{"subtitle", n.Subtitle, "", "-subtitle", true},
		{"group", n.Group, "{{.SessionID}}", "-group", false},
	}

	var args []string
	for _, tf := range fields {
		tmplStr := tf.value
		if tmplStr == "" {
			tmplStr = tf.fallback
		}
		if tmplStr == "" {
			continue
		}

		result, err := tmpl.Render(tf.name, tmplStr, tctx)
		if err != nil {
			return nil, err
		}
		if tf.required || result != "" {
			args = append(args, tf.flag, result)
		}
	}

	staticFlags := []struct{ flag, value string }{
		{"-sound", n.Sound},
		{"-open", n.Open},
		{"-execute", n.Execute},
		{"-activate", n.Activate},
		{"-sender", n.Sender},
		{"-appIcon", n.AppIcon},
		{"-contentImage", n.ContentImage},
	}
	for _, sf := range staticFlags {
		if sf.value != "" {
			args = append(args, sf.flag, sf.value)
		}
	}

	if n.IgnoreDnD {
		args = append(args, "-ignoreDnD")
	}

	return args, nil
}

func (n *TerminalNotifier) Send(ctx context.Context, notif notifier.Notification) error {
	tctx := tmpl.BuildContext(notif, n.Vars)

	args, err := n.buildArgs(tctx)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, n.Path, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("running %s: %s: %w", n.Path, string(output), err)
	}

	return nil
}

// Register adds terminal-notifier to the given plugin registry.
func Register(reg *notifier.Registry) {
	err := reg.Register("terminal-notifier", func() notifier.Notifier {
		n := &TerminalNotifier{}
		ApplyDefaults(n)

		return n
	})
	if err != nil {
		panic(err)
	}
}
