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
