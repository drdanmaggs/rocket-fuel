// Package launch handles launching processes in tmux windows after session creation.
package launch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/drdanmaggs/rocket-fuel/internal/prime"
)

// EnsureClaudeSettings creates or updates .claude/settings.json in the project
// with the SessionStart hook that runs rf prime. This is how the Integrator
// gets its context — same pattern as gastown's gt prime --hook.
func EnsureClaudeSettings(repoDir string) error {
	claudeDir := filepath.Join(repoDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		return fmt.Errorf("create .claude dir: %w", err)
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")

	// Read existing settings if they exist.
	settings := make(map[string]interface{})
	if data, err := os.ReadFile(settingsPath); err == nil {
		_ = json.Unmarshal(data, &settings)
	}

	// Ensure hooks exist.
	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = make(map[string]interface{})
	}

	// Set SessionStart hook to run rf prime.
	hooks["SessionStart"] = []map[string]interface{}{
		{
			"matcher": "",
			"hooks": []map[string]interface{}{
				{
					"type":    "command",
					"command": fmt.Sprintf("export PATH=\"$HOME/go/bin:$PATH\" && rf prime"),
				},
			},
		},
	}

	// Set PreCompact hook to re-inject context after compression.
	hooks["PreCompact"] = []map[string]interface{}{
		{
			"matcher": "",
			"hooks": []map[string]interface{}{
				{
					"type":    "command",
					"command": fmt.Sprintf("export PATH=\"$HOME/go/bin:$PATH\" && rf prime"),
				},
			},
		},
	}

	settings["hooks"] = hooks

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	return os.WriteFile(settingsPath, data, 0o644)
}

// WritePrimeContext writes the prime context to a file.
// Returns the file path.
func WritePrimeContext(repoDir string, input *prime.Input) (string, error) {
	dir := filepath.Join(repoDir, ".rocket-fuel")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}

	path := filepath.Join(dir, "integrator-context.md")
	content := prime.Build(input)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write context file: %w", err)
	}

	return path, nil
}

// IntegratorCommand returns the claude launch command for the integrator window.
func IntegratorCommand() string {
	return "claude --dangerously-skip-permissions"
}

// shellQuote wraps a string in single quotes, escaping any embedded single quotes.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
