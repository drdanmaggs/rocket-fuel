package dashboard

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddPending_createsFileAndAddsEntry(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	pm := PendingMerge{
		PR:        42,
		Title:     "Add feature X",
		Reason:    "Review pending",
		Timestamp: "2026-03-21T10:00:00Z",
	}

	if err := AddPending(tmpDir, pm); err != nil {
		t.Fatalf("AddPending failed: %v", err)
	}

	// Verify file exists.
	filePath := filepath.Join(tmpDir, ".rocket-fuel", "pending-merges.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("pending-merges.json was not created")
	}

	// Verify content.
	pending := ListPending(tmpDir)
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending merge, got %d", len(pending))
	}

	if pending[0].PR != 42 {
		t.Errorf("expected PR 42, got %d", pending[0].PR)
	}
	if pending[0].Title != "Add feature X" {
		t.Errorf("expected title 'Add feature X', got '%s'", pending[0].Title)
	}
}

func TestListPending_returnsEntries(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Add multiple entries.
	pm1 := PendingMerge{PR: 42, Title: "Feature A", Reason: "Needs review", Timestamp: "2026-03-21T10:00:00Z"}
	pm2 := PendingMerge{PR: 43, Title: "Feature B", Reason: "Needs review", Timestamp: "2026-03-21T11:00:00Z"}

	if err := AddPending(tmpDir, pm1); err != nil {
		t.Fatalf("AddPending pm1 failed: %v", err)
	}
	if err := AddPending(tmpDir, pm2); err != nil {
		t.Fatalf("AddPending pm2 failed: %v", err)
	}

	// List and verify.
	pending := ListPending(tmpDir)
	if len(pending) != 2 {
		t.Fatalf("expected 2 pending merges, got %d", len(pending))
	}

	if pending[0].PR != 42 {
		t.Errorf("expected first PR to be 42, got %d", pending[0].PR)
	}
	if pending[1].PR != 43 {
		t.Errorf("expected second PR to be 43, got %d", pending[1].PR)
	}
}

func TestRemovePending_removesSpecificEntry(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Add multiple entries.
	pm1 := PendingMerge{PR: 42, Title: "Feature A", Reason: "Needs review", Timestamp: "2026-03-21T10:00:00Z"}
	pm2 := PendingMerge{PR: 43, Title: "Feature B", Reason: "Needs review", Timestamp: "2026-03-21T11:00:00Z"}
	pm3 := PendingMerge{PR: 44, Title: "Feature C", Reason: "Needs review", Timestamp: "2026-03-21T12:00:00Z"}

	if err := AddPending(tmpDir, pm1); err != nil {
		t.Fatalf("AddPending pm1 failed: %v", err)
	}
	if err := AddPending(tmpDir, pm2); err != nil {
		t.Fatalf("AddPending pm2 failed: %v", err)
	}
	if err := AddPending(tmpDir, pm3); err != nil {
		t.Fatalf("AddPending pm3 failed: %v", err)
	}

	// Remove middle entry.
	if err := RemovePending(tmpDir, 43); err != nil {
		t.Fatalf("RemovePending failed: %v", err)
	}

	// Verify only 2 remain.
	pending := ListPending(tmpDir)
	if len(pending) != 2 {
		t.Fatalf("expected 2 pending merges after removal, got %d", len(pending))
	}

	if pending[0].PR != 42 {
		t.Errorf("expected first PR to be 42, got %d", pending[0].PR)
	}
	if pending[1].PR != 44 {
		t.Errorf("expected second PR to be 44, got %d", pending[1].PR)
	}
}

func TestListPending_returnsEmptySliceWhenFileNotExists(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Don't create any file.
	pending := ListPending(tmpDir)

	if pending == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(pending) != 0 {
		t.Fatalf("expected empty slice, got %d entries", len(pending))
	}
}
