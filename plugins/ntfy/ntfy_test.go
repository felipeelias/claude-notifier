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
		Title:   "Claude Code (myproject)",
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

func TestNtfyImplementsNotifier(t *testing.T) {
	var _ notifier.Notifier = &ntfy.Ntfy{}
}
