package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunPrimeWith_outputsOnlyRepoContextForWorkers(t *testing.T) {
	// Not parallel — calls unexported command function that may touch rootCmd.

	// Arrange: stdin contains Worker role JSON (as provided by Claude Code hooks).
	input := strings.NewReader(`{"agent_type":"worker","cwd":"/tmp/test/.worktrees/worker-42"}`)
	var out bytes.Buffer

	// Act: call the testable logic function with Worker input and captured output.
	err := runPrimeWith(input, &out)

	// Assert: no error expected.
	if err != nil {
		t.Fatalf("expected nil error for Worker role, got: %v", err)
	}

	output := out.String()

	// Workers should NOT see Board or Workers sections (those are Integrator-only).
	if strings.Contains(output, "## Board") {
		t.Errorf("Worker output should NOT contain '## Board', got:\n%s", output)
	}
	if strings.Contains(output, "## Workers") {
		t.Errorf("Worker output should NOT contain '## Workers', got:\n%s", output)
	}

	// Workers should still get Repo context (always included).
	if !strings.Contains(output, "## Repo") {
		t.Errorf("Worker output should contain '## Repo', got:\n%s", output)
	}
}

func TestRunPrimeWith_includesRepoContextForIntegrator(t *testing.T) {
	// Not parallel — calls unexported command function that may touch rootCmd.

	// Arrange: stdin contains Integrator role JSON.
	input := strings.NewReader(`{"agent_type":"integrator","cwd":"/tmp/test"}`)
	var out bytes.Buffer

	// Act: call the testable logic function with Integrator input.
	err := runPrimeWith(input, &out)

	// Assert: no error expected.
	if err != nil {
		t.Fatalf("expected nil error for Integrator role, got: %v", err)
	}

	output := out.String()

	// Integrator should always get Repo context (repo root is detected from
	// the real working directory, not the JSON cwd).
	if !strings.Contains(output, "## Repo") {
		t.Errorf("Integrator output should contain '## Repo', got:\n%s", output)
	}
}
