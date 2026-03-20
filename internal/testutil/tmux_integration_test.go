//go:build integration

package testutil

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cleanup := SetupTmuxSocket()
	code := m.Run()
	cleanup()
	os.Exit(code)
}

func TestRealTmuxNewSession(t *testing.T) {
	t.Parallel()

	tm := NewRealTmux(t)
	sessionName := "rf-test-session-" + t.Name()

	tm.NewSession(t, sessionName)

	if !tm.HasSession(t, sessionName) {
		t.Fatalf("expected session %q to exist", sessionName)
	}
}

func TestRealTmuxNewWindow(t *testing.T) {
	t.Parallel()

	tm := NewRealTmux(t)
	sessionName := "rf-test-win-" + t.Name()

	tm.NewSession(t, sessionName)
	tm.Run(t, "new-window", "-t", sessionName, "-n", "integrator")
	tm.Run(t, "new-window", "-t", sessionName, "-n", "worker-1")

	windows := tm.ListWindows(t, sessionName)

	var found int
	for _, w := range windows {
		if w == "integrator" || w == "worker-1" {
			found++
		}
	}

	if found != 2 {
		t.Errorf("expected to find integrator and worker-1 windows, got windows: %v", windows)
	}
}

func TestRealTmuxSelectWindow(t *testing.T) {
	t.Parallel()

	tm := NewRealTmux(t)
	sessionName := "rf-test-sel-" + t.Name()

	tm.NewSession(t, sessionName)
	tm.Run(t, "new-window", "-t", sessionName, "-n", "target")

	// select-window should not error
	tm.Run(t, "select-window", "-t", sessionName+":target")
}

func TestRealTmuxHasSessionReturnsFalseForMissing(t *testing.T) {
	t.Parallel()

	tm := NewRealTmux(t)

	if tm.HasSession(t, "nonexistent-session-xyz") {
		t.Error("expected HasSession to return false for nonexistent session")
	}
}
