package testutil

import (
	"strings"
	"testing"
)

func TestRunBinaryVersion(t *testing.T) {
	t.Parallel()

	out := RunBinary(t, "version")

	if !strings.HasPrefix(out, "rf") {
		t.Errorf("expected output to start with 'rf', got %q", out)
	}
}

func TestRunBinaryHelp(t *testing.T) {
	t.Parallel()

	out := RunBinary(t, "--help")

	if !strings.Contains(out, "Visionary/Integrator") {
		t.Errorf("expected help output to mention 'Visionary/Integrator', got %q", out)
	}
}

func TestRunBinaryExpectErrorOnUnknownCommand(t *testing.T) {
	t.Parallel()

	out := RunBinaryExpectError(t, "nonexistent-command")

	if !strings.Contains(out, "unknown command") {
		t.Errorf("expected 'unknown command' in error output, got %q", out)
	}
}
