package notifier

import "context"

// Notification is the data received from Claude Code's Notification hook.
type Notification struct {
	Message string `json:"message"`
	Title   string `json:"title"`
	Cwd     string `json:"cwd"`
}

// Notifier sends notifications to a specific channel.
type Notifier interface {
	Name() string
	Send(ctx context.Context, n Notification) error
}
