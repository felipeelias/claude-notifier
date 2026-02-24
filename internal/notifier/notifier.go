package notifier

import (
	"context"
	"fmt"
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

// Validate checks that notification fields are within safe size limits.
func (n Notification) Validate() error {
	type limit struct {
		name  string
		value string
		max   int
	}
	for _, check := range []limit{
		{"Message", n.Message, 4096},
		{"Title", n.Title, 256},
		{"Cwd", n.Cwd, 4096},
		{"NotificationType", n.NotificationType, 64},
		{"SessionID", n.SessionID, 128},
		{"TranscriptPath", n.TranscriptPath, 4096},
	} {
		if len(check.value) > check.max {
			return fmt.Errorf("field %s exceeds maximum length (%d > %d)", check.name, len(check.value), check.max)
		}
	}

	return nil
}

// Notifier sends notifications to a specific channel.
type Notifier interface {
	Name() string
	Send(ctx context.Context, n Notification) error
}
