package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/felipeelias/claude-notifier/internal/config"
	"github.com/felipeelias/claude-notifier/internal/dispatch"
	"github.com/felipeelias/claude-notifier/internal/notifier"
	ucli "github.com/urfave/cli/v2"
)

// Registry is the global plugin registry, populated by plugin init() functions.
var Registry = notifier.NewRegistry()

// New creates the CLI application.
func New(version string) *ucli.App {
	return &ucli.App{
		Name:    "claude-notifier",
		Usage:   "Notification dispatcher for Claude Code",
		Version: version,
		Flags: []ucli.Flag{
			&ucli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to config file",
				Value:   config.DefaultPath(),
				EnvVars: []string{"CLAUDE_NOTIFIER_CONFIG"},
			},
		},
		Action: sendAction,
		Commands: []*ucli.Command{
			initCommand(),
			testCommand(),
		},
	}
}

func loadNotifiers(configPath string) ([]notifier.Notifier, *config.Config, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, nil, err
	}

	var notifiers []notifier.Notifier
	for name, primitives := range cfg.Notifiers {
		factory, ok := Registry.All()[name]
		if !ok {
			slog.Warn("unknown notifier plugin, skipping", "name", name)
			continue
		}
		for _, prim := range primitives {
			n := factory()
			if err := cfg.Decode(prim, n); err != nil {
				return nil, nil, fmt.Errorf("decoding config for %s: %w", name, err)
			}
			notifiers = append(notifiers, n)
		}
	}

	return notifiers, cfg, nil
}

func sendAction(c *ucli.Context) error {
	const maxInputSize = 1 << 20 // 1 MiB
	var n notifier.Notification
	if err := json.NewDecoder(io.LimitReader(os.Stdin, maxInputSize)).Decode(&n); err != nil {
		slog.Error("reading notification from stdin", "error", err)
		return nil // don't fail the hook
	}

	if err := n.Validate(); err != nil {
		slog.Error("invalid notification", "error", err)
		return nil // don't fail the hook
	}

	configPath := c.String("config")
	notifiers, cfg, err := loadNotifiers(configPath)
	if err != nil {
		slog.Error("loading config", "error", err)
		return nil // don't fail the hook
	}

	ctx := c.Context
	if cfg.Global.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Global.Timeout)
		defer cancel()
	}

	if errs := dispatch.Send(ctx, notifiers, n); len(errs) > 0 {
		for _, err := range errs {
			slog.Error("sending notification", "error", err)
		}
	}

	return nil // always succeed
}

func initCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "init",
		Usage: "Create default config file",
		Action: func(c *ucli.Context) error {
			configPath := c.String("config")

			if _, err := os.Stat(configPath); err == nil {
				return fmt.Errorf("config already exists at %s", configPath)
			}

			if err := os.MkdirAll(filepath.Dir(configPath), 0750); err != nil {
				return fmt.Errorf("creating config directory: %w", err)
			}

			sample := config.SampleConfig(Registry)
			if err := os.WriteFile(configPath, []byte(sample), 0600); err != nil {
				return fmt.Errorf("writing config: %w", err)
			}

			_, _ = fmt.Fprintf(c.App.Writer, "Config created at %s\n", configPath)
			return nil
		},
	}
}

func testCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "test",
		Usage: "Send a test notification to all configured notifiers",
		Action: func(c *ucli.Context) error {
			configPath := c.String("config")
			notifiers, cfg, err := loadNotifiers(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if len(notifiers) == 0 {
				return fmt.Errorf("no notifiers configured in %s", configPath)
			}

			n := notifier.Notification{
				Message: "This is a test notification from claude-notifier",
				Title:   "claude-notifier test",
			}

			ctx := c.Context
			if cfg.Global.Timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, cfg.Global.Timeout)
				defer cancel()
			}

			if errs := dispatch.Send(ctx, notifiers, n); len(errs) > 0 {
				for _, err := range errs {
					_, _ = fmt.Fprintf(c.App.ErrWriter, "error: %s\n", err)
				}
				return errors.New("some notifiers failed")
			}

			_, _ = fmt.Fprintln(c.App.Writer, "Test notification sent successfully")
			return nil
		},
	}
}
