package main

import (
	"log/slog"
	"os"

	appcli "github.com/felipeelias/claude-notifier/internal/cli"

	// Register plugins via init functions.
	_ "github.com/felipeelias/claude-notifier/plugins/ntfy"
	_ "github.com/felipeelias/claude-notifier/plugins/terminalnotifier"
)

var version = "dev"

func main() {
	app := appcli.New(version)
	if err := app.Run(os.Args); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}
