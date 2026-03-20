//go:build integration

package tmux

import (
	"fmt"
	"os"
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
