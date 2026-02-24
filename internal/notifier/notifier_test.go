package notifier_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/felipeelias/claude-notifier/internal/notifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockNotifier struct {
	name string
	sent []notifier.Notification
}

func (m *mockNotifier) Name() string { return m.name }

func (m *mockNotifier) Send(_ context.Context, n notifier.Notification) error {
	m.sent = append(m.sent, n)

	return nil
}

func TestNotificationJSON(t *testing.T) {
	raw := `{
		"message":"Task complete",
		"title":"Claude Code",
		"cwd":"/home/user/project",
		"notification_type":"idle_prompt",
		"session_id":"abc123",
		"transcript_path":"/home/user/.claude/transcript.jsonl"
	}`
	var notif notifier.Notification
	require.NoError(t, json.Unmarshal([]byte(raw), &notif))
	assert.Equal(t, "Task complete", notif.Message)
	assert.Equal(t, "Claude Code", notif.Title)
	assert.Equal(t, "/home/user/project", notif.Cwd)
	assert.Equal(t, "idle_prompt", notif.NotificationType)
	assert.Equal(t, "abc123", notif.SessionID)
	assert.Equal(t, "/home/user/.claude/transcript.jsonl", notif.TranscriptPath)
	assert.Equal(t, "project", notif.Project())
}

func TestValidateAcceptsNormalNotification(t *testing.T) {
	notif := notifier.Notification{
		Message:          "Task complete",
		Title:            "Claude Code",
		Cwd:              "/home/user/project",
		NotificationType: "idle_prompt",
		SessionID:        "abc123",
		TranscriptPath:   "/home/user/.claude/transcript.jsonl",
	}
	assert.NoError(t, notif.Validate())
}

func TestValidateAcceptsEmptyNotification(t *testing.T) {
	n := notifier.Notification{}
	assert.NoError(t, n.Validate())
}

func TestValidateRejectsLongMessage(t *testing.T) {
	n := notifier.Notification{Message: strings.Repeat("a", 4097)}
	err := n.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Message")
}

func TestValidateRejectsLongTitle(t *testing.T) {
	n := notifier.Notification{Title: strings.Repeat("a", 257)}
	err := n.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Title")
}

func TestValidateAcceptsAtExactLimit(t *testing.T) {
	n := notifier.Notification{
		Message: strings.Repeat("a", 4096),
		Title:   strings.Repeat("a", 256),
	}
	assert.NoError(t, n.Validate())
}

func TestNotifierInterface(t *testing.T) {
	mock := &mockNotifier{name: "mock"}
	var _ notifier.Notifier = mock // compile-time interface check

	notif := notifier.Notification{Message: "hello"}
	err := mock.Send(context.Background(), notif)
	require.NoError(t, err)
	assert.Len(t, mock.sent, 1)
	assert.Equal(t, "hello", mock.sent[0].Message)
}
