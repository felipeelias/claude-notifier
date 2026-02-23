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

func (m *mockNotifier) Send(ctx context.Context, n notifier.Notification) error {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, n)
	return m.sendErr
}

func TestDispatchToMultiple(t *testing.T) {
	a := &mockNotifier{name: "a"}
	b := &mockNotifier{name: "b"}

	n := notifier.Notification{Message: "hello"}
	errs := dispatch.Send(context.Background(), []notifier.Notifier{a, b}, n)

	assert.Empty(t, errs)
	assert.Len(t, a.sent, 1)
	assert.Len(t, b.sent, 1)
}

func TestDispatchConcurrent(t *testing.T) {
	a := &mockNotifier{name: "a", delay: 50 * time.Millisecond}
	b := &mockNotifier{name: "b", delay: 50 * time.Millisecond}

	start := time.Now()
	n := notifier.Notification{Message: "hello"}
	dispatch.Send(context.Background(), []notifier.Notifier{a, b}, n)
	elapsed := time.Since(start)

	// If concurrent, should complete in ~50ms, not ~100ms
	assert.Less(t, elapsed, 90*time.Millisecond)
}

func TestDispatchCollectsErrors(t *testing.T) {
	a := &mockNotifier{name: "a", sendErr: errors.New("fail")}
	b := &mockNotifier{name: "b"}

	n := notifier.Notification{Message: "hello"}
	errs := dispatch.Send(context.Background(), []notifier.Notifier{a, b}, n)

	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "a")
	// b should still have received the notification
	assert.Len(t, b.sent, 1)
}

func TestDispatchRespectsTimeout(t *testing.T) {
	slow := &mockNotifier{name: "slow", delay: 5 * time.Second}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	n := notifier.Notification{Message: "hello"}
	errs := dispatch.Send(ctx, []notifier.Notifier{slow}, n)

	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "slow")
}
