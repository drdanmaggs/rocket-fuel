// Package session manages the Rocket Fuel tmux session lifecycle.
package session

import (
	"fmt"

	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
)

// DefaultSessionName is the tmux session name used by Rocket Fuel.
const DefaultSessionName = "rocket-fuel"

// Windows defines the tmux windows created for a Rocket Fuel session.
var Windows = []string{"integrator", "dashboard"}

// Setup creates the Rocket Fuel tmux session with all agent windows.
// Returns true if a new session was created, false if one already existed.
func Setup(tm tmux.Runner, sessionName string) (bool, error) {
	if tm.HasSession(sessionName) {
		return false, nil
	}

	if err := tm.NewSession(sessionName); err != nil {
		return false, fmt.Errorf("create session: %w", err)
	}

	for _, win := range Windows {
		if err := tm.NewWindow(sessionName, win); err != nil {
			// Clean up on partial failure.
			_ = tm.KillSession(sessionName)
			return false, fmt.Errorf("create window %q: %w", win, err)
		}
	}

	// Switch back to the first window (integrator).
	if err := tm.SelectWindow(sessionName, Windows[0]); err != nil {
		return true, fmt.Errorf("select window: %w", err)
	}

	return true, nil
}
