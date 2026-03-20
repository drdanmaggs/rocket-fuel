// Package testutil provides shared test helpers for the rocket-fuel project.
package testutil

import (
	"context"
	"os/exec"
	"strings"
	"testing"
)

// RunBinary builds and runs the rocket-fuel binary with the given args,
// returning stdout as a string. Fails the test on build or execution error.
func RunBinary(t *testing.T, args ...string) string {
	t.Helper()

	ctx := context.Background()
	binary := t.TempDir() + "/rocket-fuel"

	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	build.Dir = projectRoot(t)

	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}

	cmd := exec.CommandContext(ctx, binary, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("rocket-fuel %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}

	return strings.TrimSpace(string(out))
}

// RunBinaryExpectError builds and runs the rocket-fuel binary expecting a
// non-zero exit code. Returns combined output. Fails if the command succeeds.
func RunBinaryExpectError(t *testing.T, args ...string) string {
	t.Helper()

	ctx := context.Background()
	binary := t.TempDir() + "/rocket-fuel"

	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	build.Dir = projectRoot(t)

	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}

	cmd := exec.CommandContext(ctx, binary, args...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected rocket-fuel %s to fail, but it succeeded:\n%s", strings.Join(args, " "), out)
	}

	return strings.TrimSpace(string(out))
}

func projectRoot(t *testing.T) string {
	t.Helper()

	cmd := exec.CommandContext(context.Background(), "go", "list", "-m", "-f", "{{.Dir}}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to find project root: %v\n%s", err, out)
	}

	return strings.TrimSpace(string(out))
}
