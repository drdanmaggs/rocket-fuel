package worker

import (
	"testing"
)

func TestRouteSkillFromLabels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		labels   []string
		expected string
	}{
		{"tdd label", []string{"workflow:tdd"}, "/tdd"},
		{"bug-fix label", []string{"workflow:bug-fix"}, "/bug-fix"},
		{"epc label", []string{"workflow:epc"}, "/epc"},
		{"issue-scope label", []string{"workflow:issue-scope"}, "/issue-scope"},
		{"no workflow label", []string{"enhancement", "v0.1"}, "/epc"},
		{"empty labels", nil, "/epc"},
		{"multiple labels picks first workflow", []string{"v0.1", "workflow:tdd", "workflow:bug-fix"}, "/tdd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := RouteSkill(tt.labels)
			if got != tt.expected {
				t.Errorf("RouteSkill(%v) = %q, want %q", tt.labels, got, tt.expected)
			}
		})
	}
}

func TestBuildPrompt(t *testing.T) {
	t.Parallel()

	issue := Issue{
		Number: 42,
		Title:  "Add user login",
		Body:   "Implement OAuth2 flow",
		Labels: []string{"workflow:tdd"},
	}

	prompt := buildPrompt(issue, "/tdd")

	if got := prompt; got == "" {
		t.Fatal("expected non-empty prompt")
	}

	checks := []string{
		"#42",
		"Add user login",
		"OAuth2 flow",
		"/tdd",
		"gh pr create",
	}

	for _, check := range checks {
		if !contains(prompt, check) {
			t.Errorf("expected prompt to contain %q", check)
		}
	}
}

func TestBuildPromptWithoutBody(t *testing.T) {
	t.Parallel()

	issue := Issue{
		Number: 1,
		Title:  "Simple fix",
	}

	prompt := buildPrompt(issue, "/bug-fix")

	if contains(prompt, "Issue description") {
		t.Error("expected no 'Issue description' section when body is empty")
	}
}

func TestShellQuote(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "'hello'"},
		{"it's", "'it'\\''s'"},
		{"", "''"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := shellQuote(tt.input)
			if got != tt.expected {
				t.Errorf("shellQuote(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && // avoid false positives on empty strings
		indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
