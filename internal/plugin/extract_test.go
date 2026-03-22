package plugin_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/drdanmaggs/rocket-fuel/internal/plugin"
	"gopkg.in/yaml.v3"
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

func TestExtractPlugin_createsIntegratorAgentFileWithValidYAMLFrontmatter(t *testing.T) {
	t.Parallel()

	targetDir := t.TempDir()

	// Act
	err := plugin.ExtractPlugin(targetDir)
	if err != nil {
		t.Fatalf("ExtractPlugin() returned unexpected error: %v", err)
	}

	// Assert: agents/integrator.md exists
	agentPath := filepath.Join(targetDir, "agents", "integrator.md")
	data, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("expected agents/integrator.md to exist, got error: %v", err)
	}

	content := string(data)

	// Assert: file has YAML frontmatter between --- delimiters
	if !strings.HasPrefix(content, "---\n") {
		t.Fatal("agents/integrator.md does not start with YAML frontmatter delimiter '---'")
	}

	// Extract frontmatter between first and second ---
	rest := content[4:] // skip first "---\n"
	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		t.Fatal("agents/integrator.md missing closing YAML frontmatter delimiter '---'")
	}
	frontmatter := rest[:endIdx]

	// Parse YAML frontmatter
	var meta map[string]interface{}
	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		t.Fatalf("YAML frontmatter is not valid YAML: %v", err)
	}

	// Assert: required string fields
	for _, field := range []string{"name", "description"} {
		val, ok := meta[field]
		if !ok {
			t.Errorf("frontmatter missing required field %q", field)
			continue
		}
		str, ok := val.(string)
		if !ok || str == "" {
			t.Errorf("frontmatter field %q should be a non-empty string, got %v", field, val)
		}
	}
}

func TestExtractPlugin_createsWorkerAgentFileWithValidYAMLFrontmatter(t *testing.T) {
	t.Parallel()

	targetDir := t.TempDir()

	// Act
	err := plugin.ExtractPlugin(targetDir)
	if err != nil {
		t.Fatalf("ExtractPlugin() returned unexpected error: %v", err)
	}

	// Assert: agents/worker.md exists
	agentPath := filepath.Join(targetDir, "agents", "worker.md")
	data, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("expected agents/worker.md to exist, got error: %v", err)
	}

	content := string(data)

	// Assert: file has YAML frontmatter between --- delimiters
	if !strings.HasPrefix(content, "---\n") {
		t.Fatal("agents/worker.md does not start with YAML frontmatter delimiter '---'")
	}

	// Extract frontmatter between first and second ---
	rest := content[4:] // skip first "---\n"
	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		t.Fatal("agents/worker.md missing closing YAML frontmatter delimiter '---'")
	}
	frontmatter := rest[:endIdx]

	// Parse YAML frontmatter
	var meta map[string]interface{}
	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		t.Fatalf("YAML frontmatter is not valid YAML: %v", err)
	}

	// Assert: required string fields
	for _, field := range []string{"name", "description"} {
		val, ok := meta[field]
		if !ok {
			t.Errorf("frontmatter missing required field %q", field)
			continue
		}
		str, ok := val.(string)
		if !ok || str == "" {
			t.Errorf("frontmatter field %q should be a non-empty string, got %v", field, val)
		}
	}
}

func TestExtractPlugin_createsBoardSetupSkillWithValidFrontmatterAndColumnNames(t *testing.T) {
	t.Parallel()

	targetDir := t.TempDir()

	// Act
	err := plugin.ExtractPlugin(targetDir)
	if err != nil {
		t.Fatalf("ExtractPlugin() returned unexpected error: %v", err)
	}

	// Assert: skills/board-setup/SKILL.md exists
	skillPath := filepath.Join(targetDir, "skills", "board-setup", "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("expected skills/board-setup/SKILL.md to exist, got error: %v", err)
	}

	content := string(data)

	// Assert: file has YAML frontmatter between --- delimiters
	if !strings.HasPrefix(content, "---\n") {
		t.Fatal("skills/board-setup/SKILL.md does not start with YAML frontmatter delimiter '---'")
	}

	// Extract frontmatter between first and second ---
	rest := content[4:] // skip first "---\n"
	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		t.Fatal("skills/board-setup/SKILL.md missing closing YAML frontmatter delimiter '---'")
	}
	frontmatter := rest[:endIdx]

	// Parse YAML frontmatter
	var meta map[string]interface{}
	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		t.Fatalf("YAML frontmatter is not valid YAML: %v", err)
	}

	// Assert: required string fields
	for _, field := range []string{"name", "description"} {
		val, ok := meta[field]
		if !ok {
			t.Errorf("frontmatter missing required field %q", field)
			continue
		}
		str, ok := val.(string)
		if !ok || str == "" {
			t.Errorf("frontmatter field %q should be a non-empty string, got %v", field, val)
		}
	}

	// Assert: body contains all standard column names
	body := rest[endIdx+4:] // skip "\n---" + newline after closing delimiter
	for _, column := range []string{"Backlog", "Ready", "Scoped", "In Progress", "In Review", "Done"} {
		if !strings.Contains(body, column) {
			t.Errorf("skills/board-setup/SKILL.md body should contain column name %q", column)
		}
	}
}

func TestExtractPlugin_extractsAll17AgentDefinitions(t *testing.T) {
	t.Parallel()

	targetDir := t.TempDir()

	// Act
	err := plugin.ExtractPlugin(targetDir)
	if err != nil {
		t.Fatalf("ExtractPlugin() returned unexpected error: %v", err)
	}

	// Assert: agents/ directory contains exactly 17 .md files
	agentsDir := filepath.Join(targetDir, "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		t.Fatalf("expected agents/ directory to exist, got error: %v", err)
	}

	expectedAgents := []string{
		"codebase-scanner.md",
		"code-reviewer-bug-hunter.md",
		"code-reviewer-context-reviewer.md",
		"code-reviewer-performance-reviewer.md",
		"code-reviewer-quality-reviewer.md",
		"code-reviewer-standards-checker.md",
		"code-reviewer-test-coverage-reviewer.md",
		"code-reviewer-validator.md",
		"debt-hunter.md",
		"documentation-maintainer.md",
		"integrator.md",
		"progress-update.md",
		"tdd-implementer.md",
		"tdd-plan-reviewer.md",
		"tdd-refactorer.md",
		"tdd-test-writer.md",
		"worker.md",
	}

	// Collect only .md files
	var mdFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			mdFiles = append(mdFiles, entry.Name())
		}
	}

	if len(mdFiles) != 17 {
		t.Fatalf("expected 17 agent .md files, got %d: %v", len(mdFiles), mdFiles)
	}

	// Assert: each expected agent file exists, is non-empty, and starts with "---" or "#"
	for _, agentFile := range expectedAgents {
		agentPath := filepath.Join(agentsDir, agentFile)
		data, err := os.ReadFile(agentPath)
		if err != nil {
			t.Errorf("expected agent file %q to exist, got error: %v", agentFile, err)
			continue
		}

		if len(data) == 0 {
			t.Errorf("agent file %q is empty", agentFile)
			continue
		}

		content := string(data)
		if !strings.HasPrefix(content, "---") && !strings.HasPrefix(content, "#") {
			t.Errorf("agent file %q should start with '---' (YAML frontmatter) or '#' (markdown heading), got prefix: %q",
				agentFile, content[:min(30, len(content))])
		}
	}
}

func TestExtractPlugin_returnsErrorIfTargetDirectoryIsNotWritable(t *testing.T) {
	t.Parallel()

	// Arrange: create a read-only directory so MkdirAll fails for subdirectories
	readOnlyDir := t.TempDir()
	if err := os.Chmod(readOnlyDir, 0o555); err != nil {
		t.Fatalf("failed to set read-only permissions: %v", err)
	}
	t.Cleanup(func() {
		// Restore write permission so t.TempDir() cleanup can remove the directory
		os.Chmod(readOnlyDir, 0o755) //nolint:errcheck
	})

	unwritablePath := filepath.Join(readOnlyDir, "subdir")

	// Act
	err := plugin.ExtractPlugin(unwritablePath)

	// Assert
	if err == nil {
		t.Fatal("expected ExtractPlugin to return an error for unwritable directory, got nil")
	}
}
