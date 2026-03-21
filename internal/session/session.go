// Package session manages the Rocket Fuel tmux session lifecycle.
package session

import (
	"fmt"
	"strings"

	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
)

// DefaultSessionName is the tmux session for all Rocket Fuel windows.
const DefaultSessionName = "rf-integrator"

// Window names.
const (
	WindowIntegrator  = "integrator"
	WindowMissionCtrl = "mission-control"
)

// Setup creates the Rocket Fuel tmux session with an "integrator" window
// and a "mission-control" window. All agents (workers included) live as
// tabs in this single session so the Visionary can see everything.
// Returns true if a new session was created, false if one already existed.
func Setup(tm tmux.Runner, sessionName string) (bool, error) {
	if tm.HasSession(sessionName) {
		return false, nil
	}

	if err := tm.NewSession(sessionName); err != nil {
		return false, fmt.Errorf("create session: %w", err)
	}

	// Rename the default window (window 0) to "integrator".
	_ = tm.RenameWindow(sessionName, "0", WindowIntegrator)

	// Create mission-control window.
	if err := tm.NewWindow(sessionName, WindowMissionCtrl); err != nil {
		_ = tm.KillSession(sessionName)
		return false, fmt.Errorf("create window %q: %w", WindowMissionCtrl, err)
	}

	// Select integrator so the Visionary lands there.
	_ = tm.SelectWindow(sessionName, WindowIntegrator)

	return true, nil
}

// TeardownAll kills the session (which contains all windows — integrator,
// mission-control, and any workers).
func TeardownAll(tm tmux.Runner, sessionName string) (bool, error) {
	if !tm.HasSession(sessionName) {
		return false, nil
	}

	if err := tm.KillSession(sessionName); err != nil {
		return false, fmt.Errorf("kill session: %w", err)
	}

	return true, nil
}

// HasWindowWithPrefix checks if any window in the session has a name starting with prefix.
// Shared by worker/reap and status to match "#N:" style worker window names.
func HasWindowWithPrefix(tm tmux.Runner, sessionName, prefix string) bool {
	names, err := tm.ListWindowNames(sessionName)
	if err != nil {
		return false
	}
	for _, name := range names {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// ListWorkerWindows returns the names of worker windows in the session.
func ListWorkerWindows(tm tmux.Runner, sessionName string) []string {
	names, err := tm.ListWindowNames(sessionName)
	if err != nil {
		return nil
	}

	var workers []string
	for _, name := range names {
		// Worker windows are named "#N: title" (new) or "worker-N" (legacy).
		if strings.HasPrefix(name, "#") || strings.HasPrefix(name, "worker-") {
			workers = append(workers, name)
		}
	}
	return workers
}
