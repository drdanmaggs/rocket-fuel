package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// repoRoot returns the root directory of the current git repository.
func repoRoot() (string, error) {
	out, err := exec.CommandContext(context.Background(), "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repo: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
