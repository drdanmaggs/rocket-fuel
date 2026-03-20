// Package session manages the Rocket Fuel tmux session lifecycle.
package session

import (
	"fmt"

	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
)

// DefaultSessionName is the tmux session name used by Rocket Fuel.
const DefaultSessionName = "rocket-fuel"

// MissionControlSession is the separate tmux session for mission control.
const MissionControlSession = "rf-mission-control"

// Window names.
const (
	WindowIntegrator  = "integrator"
	WindowMissionCtrl = "mission-control"
)

// Setup creates the Rocket Fuel tmux session with a single "integrator" window.
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

// SetupMissionControl creates a separate detached tmux session for mission control.
// Returns true if created, false if already running.
func SetupMissionControl(tm tmux.Runner) (bool, error) {
	if tm.HasSession(MissionControlSession) {
		return false, nil
	}

	if err := tm.NewSession(MissionControlSession); err != nil {
		return false, fmt.Errorf("create mission control session: %w", err)
	}

	if cli, ok := tm.(*tmux.CLI); ok {
		_ = cli.RenameWindow(MissionControlSession, "0", WindowMissionCtrl)
	}

	return true, nil
}

// TeardownAll kills both the main session and mission control.
// TeardownAll kills the integrator, mission control, and all worker sessions.
func TeardownAll(tm tmux.Runner, sessionName string) (bool, error) {
	killed := false

	// Kill worker sessions (rf-worker-*).
	if cli, ok := tm.(*tmux.CLI); ok {
		for _, ws := range cli.ListSessions() {
			if len(ws) > 10 && ws[:10] == "rf-worker-" {
				_ = tm.KillSession(ws)
				killed = true
			}
		}
	}

	// Kill mission control.
	if tm.HasSession(MissionControlSession) {
		_ = tm.KillSession(MissionControlSession)
		killed = true
	}

	// Kill integrator.
	if tm.HasSession(sessionName) {
		if err := tm.KillSession(sessionName); err != nil {
			return killed, fmt.Errorf("kill session: %w", err)
		}
		killed = true
	}

	return killed, nil
}
