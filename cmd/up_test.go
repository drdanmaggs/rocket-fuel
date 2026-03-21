package cmd

import (
	"bytes"
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
