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

const (
	configDirPerms  = 0750
	configFilePerms = 0600
)

// New creates the CLI application.
func New(version string, reg *notifier.Registry) *ucli.App {
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
		Action: func(cmd *ucli.Context) error {
			return sendAction(cmd, reg)
		},
		Commands: []*ucli.Command{
			initCommand(reg),
			testCommand(reg),
		},
	}
}

func loadNotifiers(configPath string, reg *notifier.Registry) ([]notifier.Notifier, *config.Config, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, nil, err
	}

	var notifiers []notifier.Notifier
	for name, primitives := range cfg.Notifiers {
		factory, ok := reg.All()[name]
		if !ok {
			slog.Warn("unknown notifier plugin, skipping", "name", name)

			continue
		}
		for _, prim := range primitives {
			n := factory()
			err := cfg.Decode(prim, n)
			if err != nil {
				return nil, nil, fmt.Errorf("decoding config for %s: %w", name, err)
			}
			notifiers = append(notifiers, n)
		}
	}

	return notifiers, cfg, nil
}

func sendAction(cmd *ucli.Context, reg *notifier.Registry) error {
	const maxInputSize = 1 << 20 // 1 MiB
	var notif notifier.Notification

	err := json.NewDecoder(io.LimitReader(os.Stdin, maxInputSize)).Decode(&notif)
	if err != nil {
		slog.Error("reading notification from stdin", "error", err)

		return nil // don't fail the hook
	}

	err = notif.Validate()
	if err != nil {
		slog.Error("invalid notification", "error", err)

		return nil // don't fail the hook
	}

	configPath := cmd.String("config")
	notifiers, cfg, err := loadNotifiers(configPath, reg)
	if err != nil {
		slog.Error("loading config", "error", err)

		return nil // don't fail the hook
	}

	ctx := cmd.Context
	if cfg.Global.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Global.Timeout)
		defer cancel()
	}

	if errs := dispatch.Send(ctx, notifiers, notif); len(errs) > 0 {
		for _, err := range errs {
			slog.Error("sending notification", "error", err)
		}
	}

	return nil // always succeed
}

func initCommand(reg *notifier.Registry) *ucli.Command {
	return &ucli.Command{
		Name:  "init",
		Usage: "Create default config file",
		Action: func(cmd *ucli.Context) error {
			configPath := cmd.String("config")

			_, err := os.Stat(configPath)
			if err == nil {
				return fmt.Errorf("config already exists at %s", configPath)
			}

			err = os.MkdirAll(filepath.Dir(configPath), configDirPerms)
			if err != nil {
				return fmt.Errorf("creating config directory: %w", err)
			}

			sample := config.SampleConfig(reg)
			err = os.WriteFile(configPath, []byte(sample), configFilePerms)
			if err != nil {
				return fmt.Errorf("writing config: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.App.Writer, "Config created at %s\n", configPath)

			return nil
		},
	}
}

func testCommand(reg *notifier.Registry) *ucli.Command {
	return &ucli.Command{
		Name:  "test",
		Usage: "Send a test notification to all configured notifiers",
		Action: func(cmd *ucli.Context) error {
			configPath := cmd.String("config")
			notifiers, cfg, err := loadNotifiers(configPath, reg)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if len(notifiers) == 0 {
				return fmt.Errorf("no notifiers configured in %s", configPath)
			}

			notif := notifier.Notification{
				Message: "This is a test notification from claude-notifier",
				Title:   "claude-notifier test",
			}

			ctx := cmd.Context
			if cfg.Global.Timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, cfg.Global.Timeout)
				defer cancel()
			}

			if errs := dispatch.Send(ctx, notifiers, notif); len(errs) > 0 {
				for _, err := range errs {
					_, _ = fmt.Fprintf(cmd.App.ErrWriter, "error: %s\n", err)
				}

				return errors.New("some notifiers failed")
			}

			_, _ = fmt.Fprintln(cmd.App.Writer, "Test notification sent successfully")

			return nil
		},
	}
}
