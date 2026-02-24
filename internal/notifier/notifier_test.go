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

func (m *mockNotifier) Send(ctx context.Context, n notifier.Notification) error {
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
	var n notifier.Notification
	require.NoError(t, json.Unmarshal([]byte(raw), &n))
	assert.Equal(t, "Task complete", n.Message)
	assert.Equal(t, "Claude Code", n.Title)
	assert.Equal(t, "/home/user/project", n.Cwd)
	assert.Equal(t, "idle_prompt", n.NotificationType)
	assert.Equal(t, "abc123", n.SessionID)
	assert.Equal(t, "/home/user/.claude/transcript.jsonl", n.TranscriptPath)
	assert.Equal(t, "project", n.Project())
}

func TestValidateAcceptsNormalNotification(t *testing.T) {
	n := notifier.Notification{
		Message:          "Task complete",
		Title:            "Claude Code",
		Cwd:              "/home/user/project",
		NotificationType: "idle_prompt",
		SessionID:        "abc123",
		TranscriptPath:   "/home/user/.claude/transcript.jsonl",
	}
	assert.NoError(t, n.Validate())
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
	m := &mockNotifier{name: "mock"}
	var _ notifier.Notifier = m // compile-time interface check

	n := notifier.Notification{Message: "hello"}
	err := m.Send(context.Background(), n)
	require.NoError(t, err)
	assert.Len(t, m.sent, 1)
	assert.Equal(t, "hello", m.sent[0].Message)
}
