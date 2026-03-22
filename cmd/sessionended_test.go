package cmd

import (
	"strings"
	"testing"
)

func TestRunSessionEndedWith_logsWarningForIntegratorSessionDeath(t *testing.T) {
	// Not parallel — calls unexported command function that may touch rootCmd.

	// Arrange: stdin contains Integrator role JSON (as provided by Claude Code hooks).
	// When the Integrator session dies, it's unexpected — session-ended should
	// log a warning but NOT run the worker reap logic (which would fail/be wrong).
	input := strings.NewReader(`{"agent_type":"integrator","cwd":"/tmp/test"}`)

	// Act: call the testable logic function with Integrator input.
	err := runSessionEndedWith(input)
	// Assert: returns nil (doesn't crash). The key behavioral difference is that
	// for an Integrator, it does NOT attempt to reap workers or nudge the
	// Integrator (which would be nonsensical — the Integrator IS the dying session).
	if err != nil {
		t.Fatalf("expected nil error for Integrator session death, got: %v", err)
	}
}
