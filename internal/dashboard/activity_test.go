package dashboard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteActivity_createsFileAndAppends(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	repoDir := tmpDir

	// Write first line.
	err := WriteActivity(repoDir, "test message 1")
	if err != nil {
		t.Fatalf("WriteActivity: %v", err)
	}

	// Verify file was created.
	logPath := filepath.Join(repoDir, ".rocket-fuel", "activity.log")
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("activity.log not created: %v", err)
	}

	// Write second line.
	err = WriteActivity(repoDir, "test message 2")
	if err != nil {
		t.Fatalf("WriteActivity second: %v", err)
	}

	// Verify both lines are present.
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "test message 1") {
		t.Error("expected first message in log")
	}
	if !strings.Contains(text, "test message 2") {
		t.Error("expected second message in log")
	}
	if strings.Index(text, "test message 1") > strings.Index(text, "test message 2") {
		t.Error("messages out of order")
	}
}

func TestReadActivity_returnsLastNLines(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	repoDir := tmpDir

	// Write multiple lines.
	for i := 1; i <= 5; i++ {
		err := WriteActivity(repoDir, "message "+string(rune('0'+i)))
		if err != nil {
			t.Fatalf("WriteActivity: %v", err)
		}
	}

	// Read last 3 lines.
	lines := ReadActivity(repoDir, 3)

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	// Verify they're the last 3 lines.
	if !strings.Contains(lines[0], "message 3") {
		t.Error("expected message 3 in first line")
	}
	if !strings.Contains(lines[1], "message 4") {
		t.Error("expected message 4 in second line")
	}
	if !strings.Contains(lines[2], "message 5") {
		t.Error("expected message 5 in third line")
	}
}

func TestReadActivity_fileNotExist(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	repoDir := tmpDir

	// Try to read from non-existent file.
	lines := ReadActivity(repoDir, 10)

	if lines == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(lines) != 0 {
		t.Errorf("expected empty slice, got %d lines", len(lines))
	}
}
