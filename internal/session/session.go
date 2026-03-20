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

// Setup creates the Rocket Fuel tmux session.
// Window 0 is renamed to "integrator" (no orphan window).
// A background "mission-control" window is created for the mission control loop.
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

	// Create the mission-control window for the mission control loop.
	if err := tm.NewWindow(sessionName, WindowMissionCtrl); err != nil {
		_ = tm.KillSession(sessionName)
		return false, fmt.Errorf("create window %q: %w", WindowMissionCtrl, err)
	}

	// Select the integrator window so user lands there.
	if err := tm.SelectWindow(sessionName, WindowIntegrator); err != nil {
		return true, fmt.Errorf("select window: %w", err)
	}

	return true, nil
}
