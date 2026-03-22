package plugin_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/drdanmaggs/rocket-fuel/internal/plugin"
)

func TestExtractPlugin_createsPluginDirectoryStructureAtGivenPath(t *testing.T) {
	t.Parallel()

	targetDir := t.TempDir()

	// Act
	err := plugin.ExtractPlugin(targetDir)
	if err != nil {
		t.Fatalf("ExtractPlugin() returned unexpected error: %v", err)
	}

	// Assert: .claude-plugin/plugin.json exists
	pluginJSONPath := filepath.Join(targetDir, ".claude-plugin", "plugin.json")
	data, err := os.ReadFile(pluginJSONPath)
	if err != nil {
		t.Fatalf("expected .claude-plugin/plugin.json to exist, got error: %v", err)
	}

	// Assert: plugin.json is valid JSON with required fields
	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("plugin.json is not valid JSON: %v", err)
	}

	for _, field := range []string{"name", "version", "description"} {
		val, ok := manifest[field]
		if !ok {
			t.Errorf("plugin.json missing required field %q", field)
			continue
		}
		str, ok := val.(string)
		if !ok || str == "" {
			t.Errorf("plugin.json field %q should be a non-empty string, got %v", field, val)
		}
	}
}

func TestExtractPlugin_overwritesExistingFilesOnEveryCall(t *testing.T) {
	t.Parallel()

	targetDir := t.TempDir()

	// Arrange: create a stale plugin.json that should be overwritten
	pluginDir := filepath.Join(targetDir, ".claude-plugin")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}
	staleContent := []byte(`{"name": "stale"}`)
	pluginJSONPath := filepath.Join(pluginDir, "plugin.json")
	if err := os.WriteFile(pluginJSONPath, staleContent, 0o644); err != nil {
		t.Fatalf("failed to write stale plugin.json: %v", err)
	}

	// Act
	err := plugin.ExtractPlugin(targetDir)
	if err != nil {
		t.Fatalf("ExtractPlugin() returned unexpected error: %v", err)
	}

	// Assert: file was overwritten with real manifest, not stale content
	data, err := os.ReadFile(pluginJSONPath)
	if err != nil {
		t.Fatalf("expected plugin.json to exist after extract, got error: %v", err)
	}

	if string(data) == string(staleContent) {
		t.Fatal("plugin.json still contains stale content; ExtractPlugin did not overwrite")
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("overwritten plugin.json is not valid JSON: %v", err)
	}

	name, ok := manifest["name"].(string)
	if !ok || name == "" || name == "stale" {
		t.Errorf("expected plugin.json to have real manifest name, got %q", name)
	}
}
