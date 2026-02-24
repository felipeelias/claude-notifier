package tmpl_test

import (
	"testing"

	"github.com/felipeelias/claude-notifier/internal/notifier"
	"github.com/felipeelias/claude-notifier/internal/tmpl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildContext(t *testing.T) {
	notif := notifier.Notification{
		Message:          "hello",
		Title:            "title",
		Cwd:              "/home/user/myproject",
		NotificationType: "idle_prompt",
		SessionID:        "abc123",
		TranscriptPath:   "/tmp/transcript",
	}
	ctx := tmpl.BuildContext(notif, nil)

	assert.Equal(t, "hello", ctx["Message"])
	assert.Equal(t, "title", ctx["Title"])
	assert.Equal(t, "/home/user/myproject", ctx["Cwd"])
	assert.Equal(t, "myproject", ctx["Project"])
	assert.Equal(t, "idle_prompt", ctx["NotificationType"])
	assert.Equal(t, "abc123", ctx["SessionID"])
	assert.Equal(t, "/tmp/transcript", ctx["TranscriptPath"])
}

func TestBuildContextWithVars(t *testing.T) {
	notif := notifier.Notification{Message: "hi"}
	ctx := tmpl.BuildContext(notif, map[string]string{"env": "prod"})
	assert.Equal(t, "prod", ctx["Env"])
}

func TestBuildContextVarCollision(t *testing.T) {
	notif := notifier.Notification{Message: "original"}
	ctx := tmpl.BuildContext(notif, map[string]string{"message": "overridden"})
	assert.Equal(t, "original", ctx["Message"])
}

func TestBuildContextEmptyKey(t *testing.T) {
	notif := notifier.Notification{Message: "hi"}
	ctx := tmpl.BuildContext(notif, map[string]string{"": "bad"})
	assert.Equal(t, "hi", ctx["Message"])
	assert.NotContains(t, ctx, "")
}

func TestRender(t *testing.T) {
	data := map[string]string{"Name": "world"}
	result, err := tmpl.Render("test", "hello {{.Name}}", data)
	require.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestRenderMissingKeyErrors(t *testing.T) {
	data := map[string]string{"Name": "world"}
	_, err := tmpl.Render("test", "hello {{.Nonexistent}}", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rendering test template")
}

func TestRenderBadTemplate(t *testing.T) {
	_, err := tmpl.Render("test", "{{.Invalid", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rendering test template")
}
