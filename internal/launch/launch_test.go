package launch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/drdanmaggs/rocket-fuel/internal/prime"
)

func TestWritePrimeContext_createsFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	input := &prime.Input{
		IntegratorPrompt: "You are the Integrator.",
		RepoDir:          dir,
		Branch:           "main",
	}

	path, err := WritePrimeContext(dir, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPath := filepath.Join(dir, ".rocket-fuel", "integrator-context.md")
	if path != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read context file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "You are the Integrator.") {
		t.Error("expected integrator prompt in context file")
	}
	if !strings.Contains(content, "main") {
		t.Error("expected branch name in context file")
	}
}

func TestIntegratorCommand_includesPromptFile(t *testing.T) {
	t.Parallel()

	cmd := IntegratorCommand("/tmp/context.md")

	if cmd != "claude --prompt-file '/tmp/context.md'" {
		t.Errorf("unexpected command: %q", cmd)
	}
}

func TestIntegratorCommand_quotesPathWithSpaces(t *testing.T) {
	t.Parallel()

	cmd := IntegratorCommand("/home/user/my projects/.rocket-fuel/integrator-context.md")

	if !strings.Contains(cmd, "'") {
		t.Errorf("expected quoted path for spaces, got: %q", cmd)
	}
	if !strings.Contains(cmd, "my projects") {
		t.Errorf("expected full path preserved, got: %q", cmd)
	}
}
