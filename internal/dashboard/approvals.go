package dashboard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PendingMerge represents a PR awaiting approval.
type PendingMerge struct {
	PR        int    `json:"pr"`
	Title     string `json:"title"`
	Reason    string `json:"reason"`
	Timestamp string `json:"timestamp"`
}

// AddPending appends a pending merge to the queue.
func AddPending(repoDir string, pm PendingMerge) error {
	filePath := filepath.Join(repoDir, ".rocket-fuel", "pending-merges.json")

	// Ensure the directory exists.
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Read existing entries.
	existing := make([]PendingMerge, 0, 1)
	if data, err := os.ReadFile(filePath); err == nil {
		if err := json.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("parse pending merges: %w", err)
		}
	}

	// Append new entry.
	existing = append(existing, pm)

	// Write back to file.
	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal pending merges: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("write pending merges: %w", err)
	}

	return nil
}

// ListPending returns all pending merges from the queue.
// Returns an empty slice if the file doesn't exist.
func ListPending(repoDir string) []PendingMerge {
	filePath := filepath.Join(repoDir, ".rocket-fuel", "pending-merges.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []PendingMerge{}
		}
		return []PendingMerge{}
	}

	var pending []PendingMerge
	if err := json.Unmarshal(data, &pending); err != nil {
		return []PendingMerge{}
	}

	return pending
}

// RemovePending removes a pending merge by PR number.
func RemovePending(repoDir string, pr int) error {
	filePath := filepath.Join(repoDir, ".rocket-fuel", "pending-merges.json")

	// Read existing entries.
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read pending merges: %w", err)
	}

	var existing []PendingMerge
	if err := json.Unmarshal(data, &existing); err != nil {
		return fmt.Errorf("parse pending merges: %w", err)
	}

	// Remove the entry with the matching PR number.
	var filtered []PendingMerge
	for _, pm := range existing {
		if pm.PR != pr {
			filtered = append(filtered, pm)
		}
	}

	// Write back to file.
	output, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal pending merges: %w", err)
	}

	if err := os.WriteFile(filePath, output, 0o644); err != nil {
		return fmt.Errorf("write pending merges: %w", err)
	}

	return nil
}
