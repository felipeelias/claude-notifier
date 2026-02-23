package ntfy

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/felipeelias/claude-notifier/internal/cli"
	"github.com/felipeelias/claude-notifier/internal/notifier"
)

// Ntfy sends notifications via an ntfy server.
type Ntfy struct {
	URL      string `toml:"url"`
	Priority string `toml:"priority"`
	Tags     string `toml:"tags"`
	Token    string `toml:"token"`
}

func (n *Ntfy) Name() string { return "ntfy" }

func (n *Ntfy) Send(ctx context.Context, notif notifier.Notification) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.URL, strings.NewReader(notif.Message))
	if err != nil {
		return fmt.Errorf("ntfy: creating request: %w", err)
	}

	if notif.Title != "" {
		req.Header.Set("Title", notif.Title)
	}
	if n.Priority != "" {
		req.Header.Set("Priority", n.Priority)
	}
	if n.Tags != "" {
		req.Header.Set("Tags", n.Tags)
	}
	if n.Token != "" {
		req.Header.Set("Authorization", "Bearer "+n.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("ntfy: sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ntfy: server returned %s", resp.Status)
	}

	return nil
}

// SampleConfig returns example TOML configuration.
func (n *Ntfy) SampleConfig() string {
	return `[[notifiers.ntfy]]
url = "https://ntfy.sh/my-topic"
# priority = "default"
# tags = "robot"
# token = "tk_..."
`
}

func init() {
	if err := cli.Registry.Register("ntfy", func() notifier.Notifier {
		return &Ntfy{}
	}); err != nil {
		panic(err)
	}
}
