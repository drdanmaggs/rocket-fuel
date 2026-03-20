// Package notify provides attention mechanisms for surfacing tabs and sending notifications.
package notify

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
)

// Surface brings a tmux window to the foreground in iTerm2 (via tmux-CC)
// and optionally sends a macOS notification.
func Surface(tm tmux.Runner, session, window, message string) error {
	if err := tm.SelectWindow(session, window); err != nil {
		return fmt.Errorf("select window %q: %w", window, err)
	}

	if message != "" {
		_ = macOSNotify("Rocket Fuel", message)
	}

	return nil
}

// macOSNotify sends a notification via osascript.
// Best-effort — failure is not fatal.
func macOSNotify(title, message string) error {
	script := fmt.Sprintf(`display notification %q with title %q`, message, title)

	cmd := exec.CommandContext(context.Background(), "osascript", "-e", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("osascript: %w\n%s", err, out)
	}
	return nil
}
