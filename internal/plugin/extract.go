// Package plugin handles embedding and extraction of Claude plugin files.
package plugin

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed .claude-plugin/plugin.json
var pluginJSON []byte

//go:embed all:agents
var agentsFS embed.FS

//go:embed all:skills
var skillsFS embed.FS

// ExtractPlugin extracts the Claude plugin files to the given target directory.
func ExtractPlugin(targetDir string) error {
	// Create the .claude-plugin directory
	claudePluginDir := filepath.Join(targetDir, ".claude-plugin")
	if err := os.MkdirAll(claudePluginDir, 0o755); err != nil {
		return err
	}

	// Write plugin.json
	pluginJSONPath := filepath.Join(claudePluginDir, "plugin.json")
	if err := os.WriteFile(pluginJSONPath, pluginJSON, 0o644); err != nil {
		return err
	}

	// Extract agents directory
	if err := extractDir(agentsFS, "agents", filepath.Join(targetDir, "agents")); err != nil {
		return err
	}

	// Extract skills directory
	if err := extractDir(skillsFS, "skills", filepath.Join(targetDir, "skills")); err != nil {
		return err
	}

	return nil
}

// extractDir recursively extracts a directory from an embedded filesystem.
func extractDir(fsys embed.FS, src, dst string) error {
	return fs.WalkDir(fsys, src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, 0o644)
	})
}
