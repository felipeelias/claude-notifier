# claude-notifier

Notification dispatcher for [Claude Code](https://docs.anthropic.com/en/docs/claude-code) hooks. Reads JSON from stdin, fans out to all configured notification channels concurrently. Single static binary, compiled-in plugins, TOML configuration.

## Why

Claude Code has [notification hooks](https://docs.anthropic.com/en/docs/claude-code/hooks) that run a shell command when the agent needs your attention. Most people write a bash script that curls ntfy or sends a desktop notification. That works fine for one channel on one machine.

It gets annoying when you want notifications on your phone *and* your desktop, or you move to a different OS and have to rewrite the script, or you want high priority for errors but low priority for routine updates.

claude-notifier is a single binary that handles all of that:

- Sends to multiple channels from one hook (ntfy, desktop, Slack, etc.)
- Same binary and config file across Linux, macOS, and Windows
- Always exits 0 so it never breaks your hook
- `claude-notifier init`, edit the TOML, you're done

## Install

### Homebrew

```bash
brew install felipeelias/tap/claude-notifier
```

### Binary

Download from [GitHub Releases](https://github.com/felipeelias/claude-notifier/releases).

### From source

```bash
go install github.com/felipeelias/claude-notifier@latest
```

## Setup

Initialize the config file:

```bash
claude-notifier init
```

This creates `~/.config/claude-notifier/config.toml`. Edit it to configure your notification channels.

### Claude Code hook

Add to your Claude Code settings (`~/.claude/settings.json`):

```json
{
  "hooks": {
    "Notification": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "claude-notifier"
          }
        ]
      }
    ]
  }
}
```

Or install the Claude plugin which configures the hook automatically.

## Configuration

```toml
# ~/.config/claude-notifier/config.toml

[global]
# Timeout for each plugin's Send call
timeout = "10s"

# ntfy push notifications (https://docs.ntfy.sh)
[[notifiers.ntfy]]
url = "https://ntfy.sh/my-topic"
# markdown = true
# message = "{{.Message}}"
# title = "Claude Code ({{.Project}})"
# priority = ""
# tags = ""
# icon = ""
# click = ""
# attach = ""
# filename = ""
# email = ""
# delay = ""
# actions = ""
# token = ""
# username = ""
# password = ""
#
# User-defined template variables
# [notifiers.ntfy.vars]
# env = "production"

# Multiple instances of the same plugin are supported
# [[notifiers.ntfy]]
# url = "https://ntfy.sh/another-topic"
```

## Usage

The primary use case is as a Claude Code hook — it reads JSON from stdin and dispatches to all configured notifiers:

```bash
echo '{"message":"Build complete","title":"Claude Code"}' | claude-notifier
```

### Commands

| Command                     | Description                                     |
| --------------------------- | ----------------------------------------------- |
| `claude-notifier`           | Read JSON from stdin, dispatch to all notifiers |
| `claude-notifier init`      | Create default config file                      |
| `claude-notifier test`      | Send a test notification to all notifiers       |
| `claude-notifier --version` | Print version                                   |

### Flags

| Flag             | Env                      | Description                                                            |
| ---------------- | ------------------------ | ---------------------------------------------------------------------- |
| `--config`, `-c` | `CLAUDE_NOTIFIER_CONFIG` | Path to config file (default: `~/.config/claude-notifier/config.toml`) |

## Plugins

### ntfy

Sends notifications via [ntfy](https://ntfy.sh), a simple HTTP-based pub-sub service.

| Field      | Default            | Description                                                  |
| ---------- | ------------------ | ------------------------------------------------------------ |
| `url`      | (required)         | ntfy server URL including topic                              |
| `markdown` | `true`             | Enable markdown formatting (web app only)                    |
| `message`  | `{{.Message}}`     | Go template for the message body                             |
| `title`    | `Claude Code ({{.Project}})` | Go template for the notification title             |
| `priority` |                    | Message priority (`min`, `low`, `default`, `high`, `urgent`) |
| `tags`     |                    | Comma-separated emoji tags                                   |
| `icon`     |                    | Notification icon URL (JPEG/PNG)                             |
| `click`    |                    | URL opened when tapping the notification                     |
| `attach`   |                    | URL of file to attach                                        |
| `filename` |                    | Override attachment filename                                 |
| `email`    |                    | Email address for notification forwarding                    |
| `delay`    |                    | Scheduled delivery (`30m`, `2h`, `tomorrow 10am`)            |
| `actions`  |                    | Action buttons in ntfy format                                |
| `token`    |                    | Access token for authentication (Bearer)                     |
| `username` |                    | Username for basic authentication                            |
| `password` |                    | Password for basic authentication                            |
| `vars`     |                    | User-defined key-value pairs for templates                   |

#### Templates

The `message` and `title` fields are [Go templates](https://pkg.go.dev/text/template).
Available variables:

| Variable                | Source                              |
| ----------------------- | ----------------------------------- |
| `{{.Message}}`          | Notification message from Claude    |
| `{{.Title}}`            | Notification title from Claude      |
| `{{.Project}}`          | Project name (last segment of cwd)  |
| `{{.Cwd}}`              | Working directory                   |
| `{{.NotificationType}}` | `permission_prompt`, `idle_prompt`, `auth_success`, `elicitation_dialog` |
| `{{.SessionID}}`        | Claude Code session ID              |
| `{{.TranscriptPath}}`   | Path to conversation transcript     |

Custom variables defined in `[notifiers.ntfy.vars]` are also available, title-cased
(e.g., `env` becomes `{{.Env}}`).

Example:

```toml
[[notifiers.ntfy]]
url = "https://ntfy.sh/my-topic"
message = "**{{.Project}}** ({{.Env}}): {{.Message}}"
title = "{{.NotificationType}}: {{.Title}}"

[notifiers.ntfy.vars]
env = "production"
```

### Writing a plugin

1. Create `plugins/<name>/<name>.go`
2. Define a struct with `toml` tags for config fields
3. Implement the `notifier.Notifier` interface (`Name()`, `Send()`)
4. Optionally implement `config.Configurable` (`SampleConfig()`)
5. Register in `init()`: `cli.Registry.Register("name", factory)`
6. Add blank import in `main.go`: `_ "github.com/felipeelias/claude-notifier/plugins/<name>"`

## Inspiration

- [Telegraf](https://github.com/influxdata/telegraf) by InfluxData — plugin architecture, TOML config with `[[section]]` arrays, and the `init()` registry pattern
- [ntfy](https://ntfy.sh) by Philipp C. Heckel — simple, self-hostable push notifications

## License

MIT
