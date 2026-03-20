// Package selfupdate handles automatic binary updates from the source repo.
package selfupdate

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Result describes what happened during an update check.
type Result struct {
	Updated    bool
	OldVersion string
	NewVersion string
	Skipped    string // reason if skipped
}

// Check compares the running binary's version against the source repo HEAD.
// If stale, pulls latest and rebuilds the binary in-place.
// Returns the result and any error (errors are non-fatal — caller should continue).
func Check(sourceDir, currentVersion, binaryPath string) (*Result, error) {
	if sourceDir == "" {
		return &Result{Skipped: "no source directory configured"}, nil
	}

	if _, err := os.Stat(sourceDir); err != nil {
		return &Result{Skipped: "source directory not found"}, nil
	}

	// Get the latest commit on origin/main.
	headCommit, err := gitOutput(sourceDir, "rev-parse", "--short", "origin/main")
	if err != nil {
		// Try fetching first — origin/main might not be up to date.
		_ = gitRun(sourceDir, "fetch", "origin", "main", "--quiet")
		headCommit, err = gitOutput(sourceDir, "rev-parse", "--short", "origin/main")
		if err != nil {
			return &Result{Skipped: "could not read origin/main"}, nil
		}
	}

	// Compare versions.
	if currentVersion == headCommit || strings.HasPrefix(headCommit, currentVersion) || strings.HasPrefix(currentVersion, headCommit) {
		return &Result{Skipped: "already up to date"}, nil
	}

	// Pull latest.
	if err := gitRun(sourceDir, "pull", "origin", "main", "--ff-only", "--quiet"); err != nil {
		return &Result{Skipped: fmt.Sprintf("pull failed: %v", err)}, nil
	}

	// Rebuild.
	if err := goBuild(sourceDir, binaryPath); err != nil {
		return &Result{Skipped: fmt.Sprintf("build failed: %v", err)}, nil
	}

	return &Result{
		Updated:    true,
		OldVersion: currentVersion,
		NewVersion: headCommit,
	}, nil
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(context.Background(), "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func gitRun(dir string, args ...string) error {
	cmd := exec.CommandContext(context.Background(), "git", args...)
	cmd.Dir = dir
	return cmd.Run()
}

func goBuild(sourceDir, binaryPath string) error {
	cmd := exec.CommandContext(context.Background(),
		"go", "build",
		"-ldflags", fmt.Sprintf("-s -w -X github.com/drdanmaggs/rocket-fuel/cmd.Version=%s -X github.com/drdanmaggs/rocket-fuel/cmd.SourceDir=%s", shortHead(sourceDir), sourceDir),
		"-o", binaryPath,
		".",
	)
	cmd.Dir = sourceDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return nil
}

func shortHead(dir string) string {
	s, _ := gitOutput(dir, "rev-parse", "--short", "HEAD")
	return s
}
