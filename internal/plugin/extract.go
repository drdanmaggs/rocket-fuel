// Package plugin handles embedding and extraction of Claude plugin files.
package plugin

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed .claude-plugin/plugin.json
var pluginJSON []byte

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

	return nil
}
