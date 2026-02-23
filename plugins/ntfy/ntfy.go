package ntfy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/felipeelias/claude-notifier/internal/cli"
	"github.com/felipeelias/claude-notifier/internal/notifier"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// Ntfy sends notifications via an ntfy server.
type Ntfy struct {
	URL      string            `toml:"url"`
	Token    string            `toml:"token"`
	Username string            `toml:"username"`
	Password string            `toml:"password"`
	Priority string            `toml:"priority"`
	Tags     string            `toml:"tags"`
	Icon     string            `toml:"icon"`
	Click    string            `toml:"click"`
	Attach   string            `toml:"attach"`
	Filename string            `toml:"filename"`
	Email    string            `toml:"email"`
	Delay    string            `toml:"delay"`
	Actions  string            `toml:"actions"`
	Markdown bool              `toml:"markdown"`
	Message  string            `toml:"message"`
	Title    string            `toml:"title"`
	Vars     map[string]string `toml:"vars"`
}

// ApplyDefaults sets sane defaults on a new Ntfy instance.
func ApplyDefaults(n *Ntfy) {
	n.Markdown = true
	n.Message = "{{.Message}}"
	n.Title = "{{.Title}}"
}

func (n *Ntfy) Name() string { return "ntfy" }

func (n *Ntfy) Send(ctx context.Context, notif notifier.Notification) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.URL, strings.NewReader(notif.Message))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
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

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	return nil
}

// SampleConfig returns example TOML configuration.
func (n *Ntfy) SampleConfig() string {
	return `# ntfy push notifications (https://docs.ntfy.sh)
[[notifiers.ntfy]]
url = "https://ntfy.sh/my-topic"
# markdown = true
# message = "{{.Message}}"
# title = "{{.Title}}"
# priority = ""
# tags = ""
# icon = ""
# click = ""
# attach = ""
# filename = ""
# email = ""
# delay = ""
# actions = ""
# token = ""
# username = ""
# password = ""
#
# User-defined template variables
# [notifiers.ntfy.vars]
# env = "production"
`
}

func init() {
	if err := cli.Registry.Register("ntfy", func() notifier.Notifier {
		n := &Ntfy{}
		ApplyDefaults(n)
		return n
	}); err != nil {
		panic(err)
	}
}
