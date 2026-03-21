// Package projects provides a registry of recently used projects for quick access.
package projects

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Project represents a single project entry in the registry.
type Project struct {
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	LastUsed time.Time `json:"last_used"`
}

// registryPath returns the path to the global projects registry file.
func registryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".rocket-fuel", "projects.json"), nil
}

// loadRegistryFromPath reads projects from a specific registry file path.
// Internal helper for testing.
func loadRegistryFromPath(regPath string) ([]Project, error) {
	data, err := os.ReadFile(regPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Project{}, nil
		}
		return nil, fmt.Errorf("read registry: %w", err)
	}

	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("parse registry: %w", err)
	}

	// Sort by LastUsed (most recent first).
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastUsed.After(projects[j].LastUsed)
	})

	return projects, nil
}

// LoadRegistry reads the projects registry from disk.
// Returns projects sorted by LastUsed (most recent first).
// Returns an empty slice if the file doesn't exist.
func LoadRegistry() ([]Project, error) {
	regPath, err := registryPath()
	if err != nil {
		return nil, err
	}
	return loadRegistryFromPath(regPath)
}

// SaveProject adds or updates a project in the registry with the current timestamp.
func SaveProject(path, name string) error {
	regPath, err := registryPath()
	if err != nil {
		return err
	}

	// Ensure directory exists.
	if err := os.MkdirAll(filepath.Dir(regPath), 0o755); err != nil {
		return fmt.Errorf("create registry directory: %w", err)
	}

	// Load existing projects.
	projects, err := loadRegistryFromPath(regPath)
	if err != nil {
		return err
	}

	// Find and update or append new project.
	now := time.Now()
	found := false
	for i, p := range projects {
		if p.Path == path {
			projects[i].LastUsed = now
			projects[i].Name = name
			found = true
			break
		}
	}
	if !found {
		projects = append(projects, Project{
			Path:     path,
			Name:     name,
			LastUsed: now,
		})
	}

	// Sort by LastUsed before saving.
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastUsed.After(projects[j].LastUsed)
	})

	// Write back to disk.
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}

	if err := os.WriteFile(regPath, data, 0o644); err != nil {
		return fmt.Errorf("write registry: %w", err)
	}

	return nil
}

// RemoveProject removes a project from the registry by path.
func RemoveProject(path string) error {
	regPath, err := registryPath()
	if err != nil {
		return err
	}

	// Load existing projects.
	projects, err := loadRegistryFromPath(regPath)
	if err != nil {
		return err
	}

	// Filter out the project to remove.
	filtered := make([]Project, 0, len(projects))
	for _, p := range projects {
		if p.Path != path {
			filtered = append(filtered, p)
		}
	}

	// If nothing was removed, return early.
	if len(filtered) == len(projects) {
		return nil
	}

	// Write back to disk.
	data, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}

	if err := os.WriteFile(regPath, data, 0o644); err != nil {
		return fmt.Errorf("write registry: %w", err)
	}

	return nil
}

// DiscoverProjects scans homeDir for git repositories up to 2 levels deep.
// Skips hidden directories (starting with .), node_modules, and go/pkg.
// Returns projects sorted by modification time (most recent first).
func DiscoverProjects(homeDir string) []Project {
	projects := make(map[string]*Project)

	// Walk up to depth 2: homeDir -> level1 -> level2
	err := filepath.Walk(homeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir
		}

		// Calculate depth relative to homeDir
		rel, _ := filepath.Rel(homeDir, path)
		depth := strings.Count(rel, string(filepath.Separator))

		// Skip if we're past depth 2
		if depth > 2 {
			return filepath.SkipDir
		}

		// Skip hidden dirs and special directories
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") || base == "node_modules" || base == "pkg" {
			return filepath.SkipDir
		}

		// If this directory has a .git subdirectory, it's a git repo
		if info.IsDir() {
			gitPath := filepath.Join(path, ".git")
			if _, err := os.Stat(gitPath); err == nil {
				// It's a git repo
				projects[path] = &Project{
					Path:     path,
					Name:     filepath.Base(path),
					LastUsed: info.ModTime(),
				}
				// Don't recurse into git repos
				return filepath.SkipDir
			}
		}

		return nil
	})
	if err != nil {
		return []Project{}
	}

	// Convert map to slice and sort by LastUsed (most recent first)
	result := make([]Project, 0, len(projects))
	for _, p := range projects {
		result = append(result, *p)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].LastUsed.After(result[j].LastUsed)
	})

	return result
}
