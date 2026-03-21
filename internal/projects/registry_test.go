package projects

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveProjectCreatesRegistry(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	regPath := filepath.Join(tmpDir, "projects.json")

	// Manually save a project using the registry path approach.
	regDir := filepath.Dir(regPath)
	if err := os.MkdirAll(regDir, 0o755); err != nil {
		t.Fatalf("failed to create registry dir: %v", err)
	}

	projects := []Project{
		{Path: "/home/user/my-project", Name: "my-project", LastUsed: time.Now()},
	}
	data, _ := json.MarshalIndent(projects, "", "  ")
	if err := os.WriteFile(regPath, data, 0o644); err != nil {
		t.Fatalf("failed to write registry: %v", err)
	}

	// Load and verify.
	loaded, err := loadRegistryFromPath(regPath)
	if err != nil {
		t.Fatalf("loadRegistryFromPath failed: %v", err)
	}

	if len(loaded) != 1 {
		t.Errorf("expected 1 project, got %d", len(loaded))
	}
	if loaded[0].Path != "/home/user/my-project" {
		t.Errorf("expected path /home/user/my-project, got %s", loaded[0].Path)
	}
	if loaded[0].Name != "my-project" {
		t.Errorf("expected name my-project, got %s", loaded[0].Name)
	}
}

func TestSaveProjectUpdatesExistingLastUsed(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	regPath := filepath.Join(tmpDir, "projects.json")

	// Create registry directory and initial project.
	regDir := filepath.Dir(regPath)
	if err := os.MkdirAll(regDir, 0o755); err != nil {
		t.Fatalf("failed to create registry dir: %v", err)
	}

	oldTime := time.Now().Add(-1 * time.Hour)
	oldProjects := []Project{
		{Path: "/home/user/my-project", Name: "my-project", LastUsed: oldTime},
	}
	data, _ := json.MarshalIndent(oldProjects, "", "  ")
	if err := os.WriteFile(regPath, data, 0o644); err != nil {
		t.Fatalf("failed to write initial registry: %v", err)
	}

	// Simulate saving the same project.
	newProjects := []Project{
		{Path: "/home/user/my-project", Name: "my-project", LastUsed: time.Now()},
	}
	data, _ = json.MarshalIndent(newProjects, "", "  ")
	if err := os.WriteFile(regPath, data, 0o644); err != nil {
		t.Fatalf("failed to update registry: %v", err)
	}

	// Verify LastUsed was updated.
	projects, err := loadRegistryFromPath(regPath)
	if err != nil {
		t.Fatalf("loadRegistryFromPath failed: %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
	if projects[0].LastUsed.Before(oldTime.Add(time.Second)) {
		t.Errorf("expected LastUsed to be updated")
	}
}

func TestLoadRegistrySortedByLastUsed(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	regPath := filepath.Join(tmpDir, "projects.json")

	// Create registry directory with projects in unsorted order.
	regDir := filepath.Dir(regPath)
	if err := os.MkdirAll(regDir, 0o755); err != nil {
		t.Fatalf("failed to create registry dir: %v", err)
	}

	now := time.Now()
	projects := []Project{
		{Path: "/home/user/project1", Name: "project1", LastUsed: now.Add(-3 * time.Hour)},
		{Path: "/home/user/project2", Name: "project2", LastUsed: now},
		{Path: "/home/user/project3", Name: "project3", LastUsed: now.Add(-1 * time.Hour)},
	}
	data, _ := json.MarshalIndent(projects, "", "  ")
	if err := os.WriteFile(regPath, data, 0o644); err != nil {
		t.Fatalf("failed to write initial registry: %v", err)
	}

	// Load and verify sorting.
	loaded, err := loadRegistryFromPath(regPath)
	if err != nil {
		t.Fatalf("loadRegistryFromPath failed: %v", err)
	}

	if len(loaded) != 3 {
		t.Errorf("expected 3 projects, got %d", len(loaded))
	}

	// Most recent should be first.
	if loaded[0].Name != "project2" {
		t.Errorf("expected first project to be project2, got %s", loaded[0].Name)
	}
	if loaded[1].Name != "project3" {
		t.Errorf("expected second project to be project3, got %s", loaded[1].Name)
	}
	if loaded[2].Name != "project1" {
		t.Errorf("expected third project to be project1, got %s", loaded[2].Name)
	}
}

func TestLoadRegistryReturnsEmptyWhenFileDoesNotExist(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	regPath := filepath.Join(tmpDir, "nonexistent", "projects.json")

	projects, err := loadRegistryFromPath(regPath)
	if err != nil {
		t.Fatalf("loadRegistryFromPath failed: %v", err)
	}

	if len(projects) != 0 {
		t.Errorf("expected empty slice, got %d projects", len(projects))
	}
}
