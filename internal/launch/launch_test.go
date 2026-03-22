package launch

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/drdanmaggs/rocket-fuel/internal/prime"
)

func TestWritePrimeContext_createsFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	input := &prime.Input{
		RepoDir: dir,
		Branch:  "main",
	}

	path, err := WritePrimeContext(dir, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPath := filepath.Join(dir, ".rocket-fuel", "integrator-context.md")
	if path != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read context file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "main") {
		t.Error("expected branch name in context file")
	}
}

func TestIntegratorCommand_usesDangerouslySkipPermissions(t *testing.T) {
	t.Parallel()

	cmd := IntegratorCommand()

	if !strings.Contains(cmd, "claude") {
		t.Errorf("expected claude command, got: %q", cmd)
	}
	if !strings.Contains(cmd, "--dangerously-skip-permissions") {
		t.Errorf("expected --dangerously-skip-permissions, got: %q", cmd)
	}
}

func TestIntegratorCommand_referencesPluginAgent(t *testing.T) {
	t.Parallel()

	cmd := IntegratorCommand()

	if !strings.Contains(cmd, "--agent") {
		t.Errorf("expected --agent flag in command, got: %q", cmd)
	}
	if !strings.Contains(cmd, "integrator") {
		t.Errorf("expected 'integrator' agent reference in command, got: %q", cmd)
	}
}

func TestEnsureClaudeSettings_createsSettingsFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	err := EnsureClaudeSettings(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("settings file not created: %v", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Verify SessionStart hook exists with rf prime.
	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		t.Fatal("expected hooks in settings")
	}

	sessionStart, ok := hooks["SessionStart"]
	if !ok {
		t.Fatal("expected SessionStart hook")
	}

	raw, _ := json.Marshal(sessionStart)
	if !strings.Contains(string(raw), "rf prime") {
		t.Errorf("expected SessionStart hook to contain 'rf prime', got: %s", raw)
	}
}

func TestEnsureClaudeSettings_preservesExistingSettings(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write existing settings with a custom field.
	existing := `{"customField": "preserve-me"}`
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	err := EnsureClaudeSettings(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "preserve-me") {
		t.Error("expected existing settings to be preserved")
	}
	if !strings.Contains(string(data), "rf prime") {
		t.Error("expected SessionStart hook to be added")
	}
}
