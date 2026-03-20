package testutil

import (
	"testing"
)

func TestFakeTmuxRecordsCommands(t *testing.T) {
	t.Parallel()

	fake := NewFakeTmux(t)

	fake.Record("new-session", "-s", "rocket-fuel")
	fake.Record("new-window", "-t", "rocket-fuel", "-n", "integrator")

	if fake.CommandCount() != 2 {
		t.Errorf("expected 2 commands, got %d", fake.CommandCount())
	}

	last := fake.LastCommand()
	if last.Action != "new-window" {
		t.Errorf("expected last action 'new-window', got %q", last.Action)
	}
}

func TestFakeTmuxHasCommand(t *testing.T) {
	t.Parallel()

	fake := NewFakeTmux(t)
	fake.Record("select-window", "-t", "rocket-fuel:worker-1")

	if !fake.HasCommand("select-window") {
		t.Error("expected HasCommand('select-window') to be true")
	}

	if fake.HasCommand("kill-session") {
		t.Error("expected HasCommand('kill-session') to be false")
	}
}

func TestFakeTmuxLastCommandEmpty(t *testing.T) {
	t.Parallel()

	fake := NewFakeTmux(t)
	last := fake.LastCommand()

	if last.Action != "" {
		t.Errorf("expected empty action for no commands, got %q", last.Action)
	}
}
