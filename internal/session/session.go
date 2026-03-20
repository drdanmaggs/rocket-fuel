// Package session manages the Rocket Fuel tmux session lifecycle.
package session

import (
	"fmt"

	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
)

// DefaultSessionName is the tmux session name used by Rocket Fuel.
const DefaultSessionName = "rocket-fuel"

// Window names for the Rocket Fuel session.
const (
	WindowIntegrator  = "integrator"
	WindowMissionCtrl = "mission-control"
)

// Setup creates the Rocket Fuel tmux session with a single "integrator" window.
// The mission-control window should be created AFTER tmux -CC attachment
// so iTerm2 renders it as a tab (not a separate window).
// Returns true if a new session was created, false if one already existed.
func Setup(tm tmux.Runner, sessionName string) (bool, error) {
	if tm.HasSession(sessionName) {
		return false, nil
	}

	if err := tm.NewSession(sessionName); err != nil {
		return false, fmt.Errorf("create session: %w", err)
	}

	// Rename the default window (window 0) to "integrator".
	if cli, ok := tm.(*tmux.CLI); ok {
		_ = cli.RenameWindow(sessionName, "0", WindowIntegrator)
	}

	return true, nil
}
