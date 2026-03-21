//go:build integration

package projects

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestDiscoverProjects_findsRealRepos verifies that DiscoverProjects
// scans a directory and finds real git repositories.
func TestDiscoverProjects_findsRealRepos(t *testing.T) {
	// Create a temporary directory structure.
	tmpDir := t.TempDir()

	// Create a git repo at top level.
	repo1 := filepath.Join(tmpDir, "my-project")
	if err := os.Mkdir(repo1, 0o755); err != nil {
		t.Fatalf("failed to create repo1 dir: %v", err)
	}
	if err := initGitRepo(repo1); err != nil {
		t.Fatalf("failed to init git repo1: %v", err)
	}

	// Create a subdirectory with another git repo.
	subDir := filepath.Join(tmpDir, "workspace")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	repo2 := filepath.Join(subDir, "another-project")
	if err := os.Mkdir(repo2, 0o755); err != nil {
		t.Fatalf("failed to create repo2 dir: %v", err)
	}
	if err := initGitRepo(repo2); err != nil {
		t.Fatalf("failed to init git repo2: %v", err)
	}

	// Create a non-git directory (should be skipped).
	noGit := filepath.Join(tmpDir, "not-a-repo")
	if err := os.Mkdir(noGit, 0o755); err != nil {
		t.Fatalf("failed to create non-git dir: %v", err)
	}

	// Create a hidden directory (should be skipped).
	hidden := filepath.Join(tmpDir, ".hidden")
	if err := os.Mkdir(hidden, 0o755); err != nil {
		t.Fatalf("failed to create hidden dir: %v", err)
	}

	// Run DiscoverProjects.
	projects := DiscoverProjects(tmpDir)

	// Verify we found exactly 2 repos.
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d: %v", len(projects), projects)
	}

	// Verify repo names.
	names := make(map[string]bool)
	for _, p := range projects {
		names[p.Name] = true
	}

	if !names["my-project"] {
		t.Errorf("expected 'my-project' in discovered projects")
	}
	if !names["another-project"] {
		t.Errorf("expected 'another-project' in discovered projects")
	}
}

// TestDiscoverProjects_skipsNodeModules verifies that DiscoverProjects
// skips node_modules, vendor, and other common excluded directories.
func TestDiscoverProjects_skipsNodeModules(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a git repo.
	repo := filepath.Join(tmpDir, "my-project")
	if err := os.Mkdir(repo, 0o755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}
	if err := initGitRepo(repo); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create node_modules with a .git inside (should be skipped).
	nodeModules := filepath.Join(tmpDir, "node_modules")
	if err := os.Mkdir(nodeModules, 0o755); err != nil {
		t.Fatalf("failed to create node_modules: %v", err)
	}
	nodeGit := filepath.Join(nodeModules, ".git")
	if err := os.Mkdir(nodeGit, 0o755); err != nil {
		t.Fatalf("failed to create .git in node_modules: %v", err)
	}

	// Create vendor with a .git inside (should be skipped).
	vendor := filepath.Join(tmpDir, "vendor")
	if err := os.Mkdir(vendor, 0o755); err != nil {
		t.Fatalf("failed to create vendor: %v", err)
	}
	vendorGit := filepath.Join(vendor, ".git")
	if err := os.Mkdir(vendorGit, 0o755); err != nil {
		t.Fatalf("failed to create .git in vendor: %v", err)
	}

	// Run DiscoverProjects.
	projects := DiscoverProjects(tmpDir)

	// Should find only the top-level repo.
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d: %v", len(projects), projects)
	}
	if projects[0].Name != "my-project" {
		t.Errorf("expected 'my-project', got %s", projects[0].Name)
	}
}

// TestDiscoverProjects_handlesMissingDir verifies that DiscoverProjects
// returns an empty slice for a non-existent directory.
func TestDiscoverProjects_handlesMissingDir(t *testing.T) {
	projects := DiscoverProjects("/nonexistent/directory/that/does/not/exist")

	if len(projects) != 0 {
		t.Errorf("expected 0 projects for missing dir, got %d", len(projects))
	}
}

// TestDiscoverProjects_sortedByModTime verifies that discovered projects
// are sorted by modification time (most recent first).
func TestDiscoverProjects_sortedByModTime(t *testing.T) {
	tmpDir := t.TempDir()

	// Create repo1.
	repo1 := filepath.Join(tmpDir, "project-old")
	if err := os.Mkdir(repo1, 0o755); err != nil {
		t.Fatalf("failed to create repo1: %v", err)
	}
	if err := initGitRepo(repo1); err != nil {
		t.Fatalf("failed to init repo1: %v", err)
	}

	// Create repo2 (will be created after repo1, so it's newer).
	repo2 := filepath.Join(tmpDir, "project-new")
	if err := os.Mkdir(repo2, 0o755); err != nil {
		t.Fatalf("failed to create repo2: %v", err)
	}
	if err := initGitRepo(repo2); err != nil {
		t.Fatalf("failed to init repo2: %v", err)
	}

	// Run DiscoverProjects.
	projects := DiscoverProjects(tmpDir)

	// Should be sorted by mod time (newest first).
	// project-new should come before project-old.
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}

	// project-new (created later) should have a more recent mod time.
	if projects[0].Name != "project-new" {
		t.Errorf("expected first project to be 'project-new', got %s", projects[0].Name)
	}
	if projects[1].Name != "project-old" {
		t.Errorf("expected second project to be 'project-old', got %s", projects[1].Name)
	}
}

// initGitRepo initializes a git repository in the given directory.
func initGitRepo(dir string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	return cmd.Run()
}
