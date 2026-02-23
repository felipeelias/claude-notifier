package notifier_test

import (
	"context"
	"encoding/json"
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
	raw := `{"message":"Task complete","title":"Claude Code","cwd":"/home/user/project"}`
	var n notifier.Notification
	require.NoError(t, json.Unmarshal([]byte(raw), &n))
	assert.Equal(t, "Task complete", n.Message)
	assert.Equal(t, "Claude Code", n.Title)
	assert.Equal(t, "/home/user/project", n.Cwd)
}

func TestNotifierInterface(t *testing.T) {
	m := &mockNotifier{name: "mock"}
	var _ notifier.Notifier = m // compile-time interface check

	n := notifier.Notification{Message: "hello"}
	err := m.Send(context.Background(), n)
	assert.NoError(t, err)
	assert.Len(t, m.sent, 1)
	assert.Equal(t, "hello", m.sent[0].Message)
}
