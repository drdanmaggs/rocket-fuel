package selfupdate

import (
	"testing"
)

func TestCheck_skipsWhenNoSourceDir(t *testing.T) {
	t.Parallel()

	result, err := Check("", "abc123", "/tmp/rf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Updated {
		t.Error("expected skip, not update")
	}
	if result.Skipped != "no source directory configured" {
		t.Errorf("unexpected skip reason: %q", result.Skipped)
	}
}

func TestCheck_skipsWhenSourceDirMissing(t *testing.T) {
	t.Parallel()

	result, err := Check("/nonexistent/path", "abc123", "/tmp/rf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Updated {
		t.Error("expected skip, not update")
	}
	if result.Skipped != "source directory not found" {
		t.Errorf("unexpected skip reason: %q", result.Skipped)
	}
}
