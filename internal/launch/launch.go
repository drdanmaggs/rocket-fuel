// Package launch handles launching processes in tmux windows after session creation.
package launch

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/drdanmaggs/rocket-fuel/internal/prime"
)

// WindowCommand pairs a window name with the command to run in it.
type WindowCommand struct {
	Window  string
	Command string
}

// WritePrimeContext writes the prime context to a file for claude --prompt-file.
// Returns the file path.
func WritePrimeContext(repoDir string, input *prime.Input) (string, error) {
	dir := filepath.Join(repoDir, ".rocket-fuel")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}

	path := filepath.Join(dir, "integrator-context.md")
	content := prime.Build(input)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write context file: %w", err)
	}

	return path, nil
}

// IntegratorCommand returns the claude launch command for the integrator window.
func IntegratorCommand(contextFilePath string) string {
	return fmt.Sprintf("claude --prompt-file %s", contextFilePath)
}
