package dispatch

import (
	"context"
	"fmt"
	"sync"

	"github.com/felipeelias/claude-notifier/internal/notifier"
)

// Send dispatches a notification to all notifiers concurrently.
// Returns a slice of errors from notifiers that failed.
func Send(ctx context.Context, notifiers []notifier.Notifier, n notifier.Notification) []error {
	var (
		mu   sync.Mutex
		errs []error
		wg   sync.WaitGroup
	)

	for _, nr := range notifiers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := nr.Send(ctx, n); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s: %w", nr.Name(), err))
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	return errs
}
