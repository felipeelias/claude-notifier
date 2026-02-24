package ntfy_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/felipeelias/claude-notifier/internal/notifier"
	"github.com/felipeelias/claude-notifier/plugins/ntfy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNtfyName(t *testing.T) {
	p := &ntfy.Ntfy{}
	assert.Equal(t, "ntfy", p.Name())
}

func TestNtfySend(t *testing.T) {
	var gotBody string
	var gotHeaders http.Header

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		gotHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := &ntfy.Ntfy{URL: srv.URL}
	err := p.Send(context.Background(), notifier.Notification{
		Message: "Task complete",
		Title:   "Claude Code",
		Cwd:     "/home/user/myproject",
	})
	require.NoError(t, err)
	assert.Equal(t, "Task complete", gotBody)
	assert.Equal(t, "Claude Code (myproject)", gotHeaders.Get("Title"))
}

func TestNtfySendWithPriority(t *testing.T) {
	var gotHeaders http.Header

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := &ntfy.Ntfy{URL: srv.URL, Priority: "high", Tags: "robot"}
	err := p.Send(context.Background(), notifier.Notification{Message: "hi"})
	require.NoError(t, err)
	assert.Equal(t, "high", gotHeaders.Get("Priority"))
	assert.Equal(t, "robot", gotHeaders.Get("Tags"))
}

func TestNtfySendWithToken(t *testing.T) {
	var gotHeaders http.Header

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := &ntfy.Ntfy{URL: srv.URL, Token: "tk_secret"}
	err := p.Send(context.Background(), notifier.Notification{Message: "hi"})
	require.NoError(t, err)
	assert.Equal(t, "Bearer tk_secret", gotHeaders.Get("Authorization"))
}

func TestNtfySendServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	p := &ntfy.Ntfy{URL: srv.URL}
	err := p.Send(context.Background(), notifier.Notification{Message: "hi"})
	assert.Error(t, err)
}

func TestNtfyDefaults(t *testing.T) {
	// Verify the factory sets correct defaults
	p := &ntfy.Ntfy{}
	ntfy.ApplyDefaults(p)
	assert.True(t, p.Markdown)
	assert.Equal(t, "{{.Message}}", p.Message)
	assert.Equal(t, "Claude Code ({{.Project}})", p.Title)
}

func TestNtfyImplementsNotifier(t *testing.T) {
	var _ notifier.Notifier = &ntfy.Ntfy{}
}

func TestNtfyTemplateRendering(t *testing.T) {
	var gotBody string
	var gotHeaders http.Header

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		gotHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := &ntfy.Ntfy{
		URL:     srv.URL,
		Message: "**{{.Project}}**: {{.Message}}",
		Title:   "{{.NotificationType}}: {{.Title}}",
	}
	err := p.Send(context.Background(), notifier.Notification{
		Message:          "Build complete",
		Title:            "Claude Code",
		Cwd:              "/home/user/myproject",
		NotificationType: "idle_prompt",
	})
	require.NoError(t, err)
	assert.Equal(t, "**myproject**: Build complete", gotBody)
	assert.Equal(t, "idle_prompt: Claude Code", gotHeaders.Get("Title"))
}

func TestNtfyTemplateWithVars(t *testing.T) {
	var gotBody string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sender := &ntfy.Ntfy{
		URL:     srv.URL,
		Message: "{{.Env}}: {{.Message}}",
		Title:   "{{.Title}}",
		Vars:    map[string]string{"env": "production"},
	}
	err := sender.Send(context.Background(), notifier.Notification{
		Message: "Task done",
		Title:   "Claude Code",
	})
	require.NoError(t, err)
	assert.Equal(t, "production: Task done", gotBody)
}

func TestNtfyTemplateBadTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := &ntfy.Ntfy{
		URL:     srv.URL,
		Message: "{{.Invalid",
		Title:   "{{.Title}}",
	}
	err := p.Send(context.Background(), notifier.Notification{Message: "hi"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rendering message template")
}

func TestNtfyAllHeaders(t *testing.T) {
	var gotHeaders http.Header

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sender := &ntfy.Ntfy{
		URL:      srv.URL,
		Message:  "{{.Message}}",
		Title:    "{{.Title}}",
		Priority: "high",
		Tags:     "robot,warning",
		Icon:     "https://example.com/icon.png",
		Click:    "https://example.com",
		Attach:   "https://example.com/file.zip",
		Filename: "report.zip",
		Email:    "user@example.com",
		Delay:    "30m",
		Actions:  "view, Open, https://example.com",
		Markdown: true,
	}
	err := sender.Send(context.Background(), notifier.Notification{
		Message: "hi",
		Title:   "test",
	})
	require.NoError(t, err)

	assert.Equal(t, "high", gotHeaders.Get("Priority"))
	assert.Equal(t, "robot,warning", gotHeaders.Get("Tags"))
	assert.Equal(t, "https://example.com/icon.png", gotHeaders.Get("X-Icon"))
	assert.Equal(t, "https://example.com", gotHeaders.Get("X-Click"))
	assert.Equal(t, "https://example.com/file.zip", gotHeaders.Get("X-Attach"))
	assert.Equal(t, "report.zip", gotHeaders.Get("X-Filename"))
	assert.Equal(t, "user@example.com", gotHeaders.Get("X-Email"))
	assert.Equal(t, "30m", gotHeaders.Get("X-Delay"))
	assert.Equal(t, "view, Open, https://example.com", gotHeaders.Get("X-Actions"))
	assert.Equal(t, "yes", gotHeaders.Get("X-Markdown"))
}

func TestNtfyMarkdownDisabled(t *testing.T) {
	var gotHeaders http.Header

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sender := &ntfy.Ntfy{
		URL:      srv.URL,
		Message:  "{{.Message}}",
		Title:    "{{.Title}}",
		Markdown: false,
	}
	err := sender.Send(context.Background(), notifier.Notification{Message: "hi"})
	require.NoError(t, err)
	assert.Empty(t, gotHeaders.Get("X-Markdown"))
}

func TestNtfyBasicAuth(t *testing.T) {
	var gotHeaders http.Header

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sender := &ntfy.Ntfy{
		URL:      srv.URL,
		Message:  "{{.Message}}",
		Title:    "{{.Title}}",
		Username: "admin",
		Password: "secret",
	}
	err := sender.Send(context.Background(), notifier.Notification{Message: "hi"})
	require.NoError(t, err)
	// base64("admin:secret") = "YWRtaW46c2VjcmV0"
	assert.Equal(t, "Basic YWRtaW46c2VjcmV0", gotHeaders.Get("Authorization"))
}

func TestNtfyTokenOverridesBasicAuth(t *testing.T) {
	var gotHeaders http.Header

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sender := &ntfy.Ntfy{
		URL:      srv.URL,
		Message:  "{{.Message}}",
		Title:    "{{.Title}}",
		Token:    "tk_secret",
		Username: "admin",
		Password: "secret",
	}
	err := sender.Send(context.Background(), notifier.Notification{Message: "hi"})
	require.NoError(t, err)
	assert.Equal(t, "Bearer tk_secret", gotHeaders.Get("Authorization"))
}

func TestNtfyVarCollision(t *testing.T) {
	var gotBody string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// User var "message" (title-cased to "Message") should NOT override the Claude Code field
	sender := &ntfy.Ntfy{
		URL:     srv.URL,
		Message: "{{.Message}}",
		Title:   "{{.Title}}",
		Vars:    map[string]string{"message": "overridden"},
	}
	err := sender.Send(context.Background(), notifier.Notification{
		Message: "original",
	})
	require.NoError(t, err)
	assert.Equal(t, "original", gotBody)
}
