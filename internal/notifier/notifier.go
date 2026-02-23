package notifier

import (
	"context"
	"path/filepath"
)

// Notification is the data received from Claude Code's Notification hook.
type Notification struct {
	Message          string `json:"message"`
	Title            string `json:"title"`
	Cwd              string `json:"cwd"`
	NotificationType string `json:"notification_type"`
	SessionID        string `json:"session_id"`
	TranscriptPath   string `json:"transcript_path"`
}

// Project returns the last path segment of Cwd.
func (n Notification) Project() string {
	return filepath.Base(n.Cwd)
}

// Notifier sends notifications to a specific channel.
type Notifier interface {
	Name() string
	Send(ctx context.Context, n Notification) error
}
