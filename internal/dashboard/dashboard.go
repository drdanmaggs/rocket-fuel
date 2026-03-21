// Package dashboard provides a live-updating status display for the split pane.
package dashboard

import (
	"fmt"
	"strings"

	"github.com/drdanmaggs/rocket-fuel/internal/status"
)

// Data holds all data needed to render the dashboard.
type Data struct {
	Summary          *status.Summary
	ActivityLog      []string
	PendingApprovals []PendingMerge
}

// Render formats the dashboard for a narrow pane (~35 chars).
func Render(data *Data) string {
	var b strings.Builder

	_, _ = fmt.Fprintln(&b, "━━━ DASHBOARD ━━━")
	_, _ = fmt.Fprintln(&b)

	// Session state.
	if data.Summary.SessionActive {
		_, _ = fmt.Fprintln(&b, "Session: ACTIVE")
	} else {
		_, _ = fmt.Fprintln(&b, "Session: INACTIVE")
	}
	_, _ = fmt.Fprintln(&b)

	// Workers.
	if len(data.Summary.Workers) == 0 {
		_, _ = fmt.Fprintln(&b, "Workers: none")
	} else {
		_, _ = fmt.Fprintf(&b, "Workers: %d\n", len(data.Summary.Workers))
		for _, w := range data.Summary.Workers {
			icon := "✓ "
			if w.WindowOpen {
				icon = "▶ "
			}
			pr := ""
			if w.HasPR {
				pr = " [PR]"
			}
			_, _ = fmt.Fprintf(&b, " %s%s%s\n", icon, w.Name, pr)
		}
	}
	_, _ = fmt.Fprintln(&b)

	// Pending approvals.
	_, _ = fmt.Fprintln(&b, "Pending Approval:")
	if len(data.PendingApprovals) == 0 {
		_, _ = fmt.Fprintln(&b, "  (none)")
	} else {
		for _, pm := range data.PendingApprovals {
			_, _ = fmt.Fprintf(&b, "  #%d %s\n", pm.PR, pm.Reason)
		}
	}
	_, _ = fmt.Fprintln(&b)

	// Activity log.
	_, _ = fmt.Fprintln(&b, "Activity:")
	if len(data.ActivityLog) == 0 {
		_, _ = fmt.Fprintln(&b, "  (none)")
	} else {
		for _, entry := range data.ActivityLog {
			_, _ = fmt.Fprintf(&b, "  %s\n", entry)
		}
	}

	return b.String()
}
