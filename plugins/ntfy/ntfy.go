package ntfy

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
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
	n.Title = "Claude Code ({{.Project}})"
}

func (n *Ntfy) Name() string { return "ntfy" }

// buildTemplateContext creates a flat map from notification fields + user vars.
func buildTemplateContext(notif notifier.Notification, vars map[string]string) map[string]string {
	ctx := map[string]string{
		"Message":          notif.Message,
		"Title":            notif.Title,
		"Cwd":              notif.Cwd,
		"Project":          notif.Project(),
		"NotificationType": notif.NotificationType,
		"SessionID":        notif.SessionID,
		"TranscriptPath":   notif.TranscriptPath,
	}
	// User vars (title-cased). Claude Code fields take precedence.
	for k, v := range vars {
		if k == "" {
			continue
		}
		key := strings.ToUpper(k[:1]) + k[1:]
		if _, exists := ctx[key]; !exists {
			ctx[key] = v
		}
	}
	return ctx
}

func renderTemplate(name, tmpl string, data map[string]string) (string, error) {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("rendering %s template: %w", name, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("rendering %s template: %w", name, err)
	}
	return buf.String(), nil
}

func (n *Ntfy) Send(ctx context.Context, notif notifier.Notification) error {
	tctx := buildTemplateContext(notif, n.Vars)

	msgTmpl := n.Message
	if msgTmpl == "" {
		msgTmpl = "{{.Message}}"
	}
	body, err := renderTemplate("message", msgTmpl, tctx)
	if err != nil {
		return err
	}

	titleTmpl := n.Title
	if titleTmpl == "" {
		titleTmpl = "Claude Code ({{.Project}})"
	}
	title, err := renderTemplate("title", titleTmpl, tctx)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.URL, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if title != "" {
		req.Header.Set("Title", title)
	}
	if n.Priority != "" {
		req.Header.Set("Priority", n.Priority)
	}
	if n.Tags != "" {
		req.Header.Set("Tags", n.Tags)
	}
	if n.Icon != "" {
		req.Header.Set("X-Icon", n.Icon)
	}
	if n.Click != "" {
		req.Header.Set("X-Click", n.Click)
	}
	if n.Attach != "" {
		req.Header.Set("X-Attach", n.Attach)
	}
	if n.Filename != "" {
		req.Header.Set("X-Filename", n.Filename)
	}
	if n.Email != "" {
		req.Header.Set("X-Email", n.Email)
	}
	if n.Delay != "" {
		req.Header.Set("X-Delay", n.Delay)
	}
	if n.Actions != "" {
		req.Header.Set("X-Actions", n.Actions)
	}
	if n.Markdown {
		req.Header.Set("X-Markdown", "yes")
	}
	// Auth: token takes precedence over username/password
	if n.Token != "" {
		req.Header.Set("Authorization", "Bearer "+n.Token)
	} else if n.Username != "" && n.Password != "" {
		creds := base64.StdEncoding.EncodeToString([]byte(n.Username + ":" + n.Password))
		req.Header.Set("Authorization", "Basic "+creds)
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
# title = "Claude Code ({{.Project}})"
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
