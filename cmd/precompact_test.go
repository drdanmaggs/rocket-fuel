package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunPrecompactWith_returnsNilWithoutOutputForWorker(t *testing.T) {
	// Not parallel — calls unexported command function that may touch rootCmd.

	// Arrange: stdin contains Worker role JSON (as provided by Claude Code hooks).
	input := strings.NewReader(`{"agent_type":"worker","cwd":"/tmp/test/.worktrees/worker-42"}`)
	var out bytes.Buffer

	// Act: call the testable logic function with Worker input and captured output.
	err := runPrecompactWith(input, &out)
	// Assert: no error expected — Worker path is a no-op placeholder.
	if err != nil {
		t.Fatalf("expected nil error for Worker role, got: %v", err)
	}

	// Assert: no output — Worker doesn't prime context.
	if out.Len() != 0 {
		t.Errorf("expected empty output for Worker role, got: %q", out.String())
	}
}

func TestRunPrecompactWith_returnsPrimeContextForIntegrator(t *testing.T) {
	// Not parallel — calls unexported command function that may touch rootCmd.

	// Arrange: stdin contains Integrator role JSON (as provided by Claude Code hooks).
	input := strings.NewReader(`{"agent_type":"integrator","cwd":"/tmp/test"}`)
	var out bytes.Buffer

	// Act: call the testable logic function with Integrator input and captured output.
	err := runPrecompactWith(input, &out)
	// Assert: no error expected.
	if err != nil {
		t.Fatalf("expected nil error for Integrator role, got: %v", err)
	}

	output := out.String()

	// Integrator PreCompact should re-prime context, which always includes Repo section.
	if !strings.Contains(output, "## Repo") {
		t.Errorf("Integrator PreCompact output should contain '## Repo', got:\n%s", output)
	}
}
