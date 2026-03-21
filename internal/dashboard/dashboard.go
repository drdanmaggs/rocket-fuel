// Package dashboard provides a live-updating status display for the split pane.
package dashboard

import (
	"fmt"
	"strings"

	"github.com/drdanmaggs/rocket-fuel/internal/status"
)

// Render formats the dashboard for a narrow pane (~35 chars).
func Render(s *status.Summary) string {
	var b strings.Builder

	_, _ = fmt.Fprintln(&b, "━━━ DASHBOARD ━━━")
	_, _ = fmt.Fprintln(&b)

	// Session state.
	if s.SessionActive {
		_, _ = fmt.Fprintln(&b, "Session: ACTIVE")
	} else {
		_, _ = fmt.Fprintln(&b, "Session: INACTIVE")
	}
	_, _ = fmt.Fprintln(&b)

	// Workers.
	if len(s.Workers) == 0 {
		_, _ = fmt.Fprintln(&b, "Workers: none")
	} else {
		_, _ = fmt.Fprintf(&b, "Workers: %d\n", len(s.Workers))
		for _, w := range s.Workers {
			icon := "  "
			if w.WindowOpen {
				icon = "▶ "
			} else {
				icon = "✓ "
			}
			pr := ""
			if w.HasPR {
				pr = " [PR]"
			}
			_, _ = fmt.Fprintf(&b, " %s%s%s\n", icon, w.Name, pr)
		}
	}

	return b.String()
}
