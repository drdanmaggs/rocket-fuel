package cmd

import (
	"strings"
	"testing"
)

func TestRunRecordActivityWith_isNoOpForIntegrator(t *testing.T) {
	// Not parallel — calls unexported command function that may touch rootCmd.

	// Arrange: stdin contains Integrator role JSON (as provided by Claude Code hooks).
	input := strings.NewReader(`{"agent_type":"integrator","cwd":"/tmp/test"}`)

	// Act: call the testable logic function with Integrator input.
	// For the Integrator role, record-activity should short-circuit immediately
	// since activity tracking only applies to Workers.
	err := runRecordActivityWith(input)

	// Assert: Integrator should return nil without performing any activity recording.
	if err != nil {
		t.Fatalf("expected nil error for Integrator role (no-op), got: %v", err)
	}
}
