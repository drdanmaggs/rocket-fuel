//go:build integration

package tmux

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

var testSocket string

func TestMain(m *testing.M) {
	testSocket = fmt.Sprintf("rf-tmux-test-%d", os.Getpid())
	code := m.Run()
	// Best-effort cleanup of the isolated tmux server.
	cli := NewWithSocket(testSocket)
	_ = cli.run("kill-server")
	os.Exit(code)
}

func TestNewSessionAndHasSession(t *testing.T) {
	t.Parallel()

	cli := NewWithSocket(testSocket)
	name := "rf-test-" + t.Name()

	if err := cli.NewSession(name); err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	t.Cleanup(func() { _ = cli.KillSession(name) })

	if !cli.HasSession(name) {
		t.Fatalf("expected session %q to exist", name)
	}
}

func TestHasSessionReturnsFalseForMissing(t *testing.T) {
	t.Parallel()

	cli := NewWithSocket(testSocket)

	if cli.HasSession("nonexistent-session-xyz") {
		t.Error("expected HasSession to return false for nonexistent session")
	}
}

func TestNewWindowCreatesWindow(t *testing.T) {
	t.Parallel()

	cli := NewWithSocket(testSocket)
	session := "rf-test-win-" + t.Name()

	if err := cli.NewSession(session); err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	t.Cleanup(func() { _ = cli.KillSession(session) })

	if err := cli.NewWindow(session, "integrator"); err != nil {
		t.Fatalf("NewWindow failed: %v", err)
	}

	if err := cli.NewWindow(session, "dashboard"); err != nil {
		t.Fatalf("NewWindow failed: %v", err)
	}
}

func TestSelectWindow(t *testing.T) {
	t.Parallel()

	cli := NewWithSocket(testSocket)
	session := "rf-test-sel-" + t.Name()

	if err := cli.NewSession(session); err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	t.Cleanup(func() { _ = cli.KillSession(session) })

	if err := cli.NewWindow(session, "target"); err != nil {
		t.Fatalf("NewWindow failed: %v", err)
	}

	if err := cli.SelectWindow(session, "target"); err != nil {
		t.Fatalf("SelectWindow failed: %v", err)
	}
}

func TestKillSession(t *testing.T) {
	t.Parallel()

	cli := NewWithSocket(testSocket)
	session := "rf-test-kill-" + t.Name()

	if err := cli.NewSession(session); err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}

	if err := cli.KillSession(session); err != nil {
		t.Fatalf("KillSession failed: %v", err)
	}

	if cli.HasSession(session) {
		t.Error("expected session to be gone after KillSession")
	}
}

func TestSplitPane_createsPane(t *testing.T) {
	t.Parallel()

	cli := NewWithSocket(testSocket)
	session := "rf-test-split-" + t.Name()

	if err := cli.NewSession(session); err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	t.Cleanup(func() { _ = cli.KillSession(session) })

	window := "main"
	if err := cli.NewWindow(session, window); err != nil {
		t.Fatalf("NewWindow failed: %v", err)
	}

	// Split the pane horizontally with a shell command using -l (lines) instead of -p (percent)
	// since -p has compatibility issues with some tmux versions.
	if err := cli.Run("split-window", "-h", "-t", session+":"+window, "-l", "10", "sleep 30"); err != nil {
		t.Fatalf("split-window failed: %v", err)
	}

	// Verify 2 panes exist via list-panes.
	panes, err := cli.output("list-panes", "-t", session+":"+window)
	if err != nil {
		t.Fatalf("list-panes failed: %v", err)
	}

	paneCount := 0
	for _, line := range strings.Split(panes, "\n") {
		if line != "" {
			paneCount++
		}
	}

	if paneCount != 2 {
		t.Errorf("expected 2 panes after split, got %d", paneCount)
	}
}

func TestListWindowNames(t *testing.T) {
	t.Parallel()

	cli := NewWithSocket(testSocket)
	session := "rf-test-list-win-" + t.Name()

	if err := cli.NewSession(session); err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	t.Cleanup(func() { _ = cli.KillSession(session) })

	// Create multiple windows.
	windows := []string{"alpha", "beta", "gamma"}
	for _, w := range windows {
		if err := cli.NewWindow(session, w); err != nil {
			t.Fatalf("NewWindow %q failed: %v", w, err)
		}
	}

	names, err := cli.ListWindowNames(session)
	if err != nil {
		t.Fatalf("ListWindowNames failed: %v", err)
	}

	// Should have the default window (0 renamed) plus the 3 we created.
	// But since we don't rename 0, it will be "0", plus alpha, beta, gamma = 4 windows.
	// Actually NewSession creates window 0, so we have 1 + 3 = 4 windows.
	if len(names) < 3 {
		t.Errorf("expected at least 3 windows, got %d: %v", len(names), names)
	}

	// Verify all created windows are in the list.
	for _, w := range windows {
		found := false
		for _, n := range names {
			if n == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected window %q in list, got: %v", w, names)
		}
	}
}

func TestListSessions(t *testing.T) {
	t.Parallel()

	cli := NewWithSocket(testSocket)

	// Create a few sessions.
	sessions := []string{"rf-test-sess1-" + t.Name(), "rf-test-sess2-" + t.Name(), "rf-test-sess3-" + t.Name()}
	for _, s := range sessions {
		if err := cli.NewSession(s); err != nil {
			t.Fatalf("NewSession %q failed: %v", s, err)
		}
		t.Cleanup(func(sessionName string) func() {
			return func() { _ = cli.KillSession(sessionName) }
		}(s))
	}

	names := cli.ListSessions()

	// Verify all created sessions are in the list.
	for _, s := range sessions {
		found := false
		for _, n := range names {
			if n == s {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected session %q in list, got: %v", s, names)
		}
	}
}
