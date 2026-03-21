package dashboard

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// WriteActivity appends a timestamped line to the activity log.
func WriteActivity(repoDir, message string) error {
	logPath := filepath.Join(repoDir, ".rocket-fuel", "activity.log")

	// Ensure the directory exists.
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Open file in append mode, creating if it doesn't exist.
	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open activity log: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Write timestamped line.
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] %s\n", timestamp, message)
	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("write activity log: %w", err)
	}

	return nil
}

// ReadActivity reads the last N lines from the activity log.
// Returns an empty slice if the file doesn't exist.
func ReadActivity(repoDir string, lines int) []string {
	logPath := filepath.Join(repoDir, ".rocket-fuel", "activity.log")

	f, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}
		}
		return []string{}
	}
	defer func() { _ = f.Close() }()

	var allLines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	// Return last N lines.
	if len(allLines) <= lines {
		return allLines
	}
	return allLines[len(allLines)-lines:]
}
