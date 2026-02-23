package tmpl

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/felipeelias/claude-notifier/internal/notifier"
)

// BuildContext creates a flat map from notification fields + user vars.
// User var keys are title-cased. Claude Code fields take precedence on collision.
func BuildContext(notif notifier.Notification, vars map[string]string) map[string]string {
	ctx := map[string]string{
		"Message":          notif.Message,
		"Title":            notif.Title,
		"Cwd":              notif.Cwd,
		"Project":          notif.Project(),
		"NotificationType": notif.NotificationType,
		"SessionID":        notif.SessionID,
		"TranscriptPath":   notif.TranscriptPath,
	}
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

// Render parses and executes a Go text/template against the given data map.
func Render(name, tmplStr string, data map[string]string) (string, error) {
	t, err := template.New(name).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("rendering %s template: %w", name, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("rendering %s template: %w", name, err)
	}
	return buf.String(), nil
}
