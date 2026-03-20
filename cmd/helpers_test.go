package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseIssueRef_plainNumber(t *testing.T) {
	t.Parallel()
	n, err := parseIssueRef("42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 42 {
		t.Errorf("expected 42, got %d", n)
	}
}

func TestParseIssueRef_withHash(t *testing.T) {
	t.Parallel()
	n, err := parseIssueRef("#42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 42 {
		t.Errorf("expected 42, got %d", n)
	}
}

func TestParseIssueRef_githubURL(t *testing.T) {
	t.Parallel()
	n, err := parseIssueRef("https://github.com/owner/repo/issues/42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 42 {
		t.Errorf("expected 42, got %d", n)
	}
}

func TestParseIssueRef_invalidInput(t *testing.T) {
	t.Parallel()
	_, err := parseIssueRef("not-a-number")
	if err == nil {
		t.Fatal("expected error for non-numeric input")
	}
}

func TestParseProjectRef_userURL(t *testing.T) {
	t.Parallel()
	owner, num, err := parseProjectRef("https://github.com/users/drdanmaggs/projects/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if owner != "drdanmaggs" {
		t.Errorf("expected owner 'drdanmaggs', got %q", owner)
	}
	if num != 1 {
		t.Errorf("expected project 1, got %d", num)
	}
}

func TestParseProjectRef_orgURL(t *testing.T) {
	t.Parallel()
	owner, num, err := parseProjectRef("https://github.com/orgs/myorg/projects/5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if owner != "myorg" {
		t.Errorf("expected owner 'myorg', got %q", owner)
	}
	if num != 5 {
		t.Errorf("expected project 5, got %d", num)
	}
}

func TestParseProjectRef_trailingSlash(t *testing.T) {
	t.Parallel()
	_, num, err := parseProjectRef("https://github.com/users/owner/projects/3/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if num != 3 {
		t.Errorf("expected 3, got %d", num)
	}
}

func TestParseProjectRef_invalidURL(t *testing.T) {
	t.Parallel()
	_, _, err := parseProjectRef("not-a-url")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestParseProjectRef_invalidNumber(t *testing.T) {
	t.Parallel()
	_, _, err := parseProjectRef("https://github.com/users/owner/projects/abc")
	if err == nil {
		t.Fatal("expected error for non-numeric project number")
	}
}

func TestLoadMaxWorkers_defaultWhenNoConfig(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	got := loadMaxWorkers(dir)
	if got != defaultMaxWorkers {
		t.Errorf("expected default %d, got %d", defaultMaxWorkers, got)
	}
}

func TestLoadMaxWorkers_readsConfig(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".rocket-fuel")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"max_workers": 5}`), 0o644); err != nil {
		t.Fatal(err)
	}

	got := loadMaxWorkers(dir)
	if got != 5 {
		t.Errorf("expected 5, got %d", got)
	}
}

func TestLoadMaxWorkers_defaultOnInvalidJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".rocket-fuel")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`not json`), 0o644); err != nil {
		t.Fatal(err)
	}

	got := loadMaxWorkers(dir)
	if got != defaultMaxWorkers {
		t.Errorf("expected default %d, got %d", defaultMaxWorkers, got)
	}
}

func TestLoadMaxWorkers_defaultOnZeroValue(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".rocket-fuel")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"max_workers": 0}`), 0o644); err != nil {
		t.Fatal(err)
	}

	got := loadMaxWorkers(dir)
	if got != defaultMaxWorkers {
		t.Errorf("expected default %d for zero value, got %d", defaultMaxWorkers, got)
	}
}
