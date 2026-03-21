package cmd

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

// NOTE: selfUpdate reads package-level SourceDir and Version vars.
// These tests must NOT be parallel (they mutate package state).

func TestSelfUpdate_BadSourceDir_SilentSkip(t *testing.T) {
	// Not parallel — mutates package-level SourceDir.

	origSourceDir := SourceDir
	origVersion := Version
	defer func() {
		SourceDir = origSourceDir
		Version = origVersion
	}()

	SourceDir = "/nonexistent/path/that/does/not/exist"
	Version = "test"

	var buf bytes.Buffer
	selfUpdate(&buf)

	// A non-existent source dir is a normal skip ("source directory not found").
	// Skips must be silent — no output.
	if buf.Len() != 0 {
		t.Errorf("expected no output for skipped update, got %q", buf.String())
	}
}

func TestSelfUpdate_EmptySourceDir_SilentSkip(t *testing.T) {
	// Not parallel — mutates package-level SourceDir.

	origSourceDir := SourceDir
	origVersion := Version
	defer func() {
		SourceDir = origSourceDir
		Version = origVersion
	}()

	SourceDir = ""
	Version = "test"

	var buf bytes.Buffer
	selfUpdate(&buf)

	// Empty source dir is a normal skip ("no source directory configured").
	// Skips must be silent — no output.
	if buf.Len() != 0 {
		t.Errorf("expected no output for skipped update, got %q", buf.String())
	}
}

func TestCheckTmux_returnsNilWhenInstalled(t *testing.T) {
	t.Parallel()

	// tmux is installed in the dev/CI environment, so this should pass.
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not installed — cannot test happy path")
	}

	if err := checkTmux(); err != nil {
		t.Errorf("expected nil error when tmux is installed, got: %v", err)
	}
}

func TestCheckTmux_returnsErrorWithInstallInstructions(t *testing.T) {
	// Not parallel — t.Setenv mutates process environment.
	t.Setenv("PATH", t.TempDir())

	err := checkTmux()
	if err == nil {
		t.Fatal("expected error when tmux is not on PATH")
	}

	msg := err.Error()
	if !strings.Contains(msg, "tmux is required but not installed") {
		t.Errorf("expected error to mention tmux is required, got: %q", msg)
	}
	if !strings.Contains(msg, "brew install tmux") {
		t.Errorf("expected macOS install instructions, got: %q", msg)
	}
	if !strings.Contains(msg, "apt install tmux") {
		t.Errorf("expected Linux install instructions, got: %q", msg)
	}
}
