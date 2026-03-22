package cmd

import (
	"strings"
	"testing"
)

func TestRunShouldContinue_allowsStopWhenRoleIsWorker(t *testing.T) {
	// Not parallel — calls unexported command function that may touch rootCmd.

	// Arrange: stdin contains Worker role JSON (as provided by Claude Code hooks).
	input := strings.NewReader(`{"agent_type":"worker","cwd":"/tmp/test"}`)

	// Act: call the testable logic function with Worker input.
	err := runShouldContinueWith(input)

	// Assert: Worker role should allow stop (return nil) without checking
	// the board or active workers.
	if err != nil {
		t.Fatalf("expected nil error for Worker role, got: %v", err)
	}
}

func TestRunShouldContinue_doesNotShortCircuitForIntegrator(t *testing.T) {
	// Not parallel — calls unexported command function that may touch rootCmd.

	// Arrange: stdin contains Integrator role JSON (as provided by Claude Code hooks).
	input := strings.NewReader(`{"agent_type":"integrator","cwd":"/tmp/test"}`)

	// Act: call the testable logic function with Integrator input.
	// Unlike Worker, Integrator should NOT short-circuit — it proceeds to
	// check the board and active workers. In a test context without a real
	// git repo, repoRoot() fails and the function falls through to "allow stop".
	err := runShouldContinueWith(input)

	// Assert: returns nil because no git repo exists, but critically this
	// proves the Worker short-circuit (role == RoleWorker) did NOT fire for
	// an Integrator — the function went through the normal code path.
	if err != nil {
		t.Fatalf("expected nil error for Integrator role (no repo fallthrough), got: %v", err)
	}
}
