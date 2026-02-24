package dispatch_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/felipeelias/claude-notifier/internal/dispatch"
	"github.com/felipeelias/claude-notifier/internal/notifier"
	"github.com/stretchr/testify/assert"
)

type mockNotifier struct {
	name    string
	mu      sync.Mutex
	sent    []notifier.Notification
	sendErr error
	delay   time.Duration
}

func (m *mockNotifier) Name() string { return m.name }

func (m *mockNotifier) Send(ctx context.Context, notif notifier.Notification) error {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, notif)

	return m.sendErr
}

func TestDispatchToMultiple(t *testing.T) {
	first := &mockNotifier{name: "first"}
	second := &mockNotifier{name: "second"}

	notif := notifier.Notification{Message: "hello"}
	errs := dispatch.Send(context.Background(), []notifier.Notifier{first, second}, notif)

	assert.Empty(t, errs)
	assert.Len(t, first.sent, 1)
	assert.Len(t, second.sent, 1)
}

func TestDispatchConcurrent(t *testing.T) {
	first := &mockNotifier{name: "first", delay: 50 * time.Millisecond}
	second := &mockNotifier{name: "second", delay: 50 * time.Millisecond}

	start := time.Now()
	notif := notifier.Notification{Message: "hello"}
	dispatch.Send(context.Background(), []notifier.Notifier{first, second}, notif)
	elapsed := time.Since(start)

	// If concurrent, should complete in ~50ms, not ~100ms
	assert.Less(t, elapsed, 90*time.Millisecond)
}

func TestDispatchCollectsErrors(t *testing.T) {
	failing := &mockNotifier{name: "failing", sendErr: errors.New("fail")}
	working := &mockNotifier{name: "working"}

	notif := notifier.Notification{Message: "hello"}
	errs := dispatch.Send(context.Background(), []notifier.Notifier{failing, working}, notif)

	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "failing")
	// working should still have received the notification
	assert.Len(t, working.sent, 1)
}

func TestDispatchRespectsTimeout(t *testing.T) {
	slow := &mockNotifier{name: "slow", delay: 5 * time.Second}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	notif := notifier.Notification{Message: "hello"}
	errs := dispatch.Send(ctx, []notifier.Notifier{slow}, notif)

	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "slow")
}
