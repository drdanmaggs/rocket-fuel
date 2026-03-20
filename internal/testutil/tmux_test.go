package testutil

import (
	"testing"
)

// RecordingTmux tests (unit — no real tmux needed)

func TestRecordingTmuxRecordsCommands(t *testing.T) {
	t.Parallel()

	rec := NewRecordingTmux()

	rec.Record("new-session", "-s", "rocket-fuel")
	rec.Record("new-window", "-t", "rocket-fuel", "-n", "integrator")

	if rec.CommandCount() != 2 {
		t.Errorf("expected 2 commands, got %d", rec.CommandCount())
	}

	last := rec.LastCommand()
	if last.Action != "new-window" {
		t.Errorf("expected last action 'new-window', got %q", last.Action)
	}
}

func TestRecordingTmuxHasCommand(t *testing.T) {
	t.Parallel()

	rec := NewRecordingTmux()
	rec.Record("select-window", "-t", "rocket-fuel:worker-1")

	if !rec.HasCommand("select-window") {
		t.Error("expected HasCommand('select-window') to be true")
	}

	if rec.HasCommand("kill-session") {
		t.Error("expected HasCommand('kill-session') to be false")
	}
}

func TestRecordingTmuxLastCommandEmpty(t *testing.T) {
	t.Parallel()

	rec := NewRecordingTmux()
	last := rec.LastCommand()

	if last.Action != "" {
		t.Errorf("expected empty action for no commands, got %q", last.Action)
	}
}
