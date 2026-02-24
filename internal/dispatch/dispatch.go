package dispatch

import (
	"context"
	"fmt"
	"sync"

	"github.com/felipeelias/claude-notifier/internal/notifier"
)

// Send dispatches a notification to all notifiers concurrently.
// Returns a slice of errors from notifiers that failed.
func Send(ctx context.Context, notifiers []notifier.Notifier, notif notifier.Notification) []error {
	var (
		mu   sync.Mutex
		errs []error
		wg   sync.WaitGroup
	)

	for _, dest := range notifiers {
		wg.Go(func() {
			err := dest.Send(ctx, notif)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s: %w", dest.Name(), err))
				mu.Unlock()
			}
		})
	}

	wg.Wait()

	return errs
}
