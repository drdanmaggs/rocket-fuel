package project

import (
	"testing"
)

func TestDiscover_parsesProjectFromResponse(t *testing.T) {
	t.Parallel()

	response := []byte(`{"data":{"repository":{"projectsV2":{"nodes":[{"id":"PVT_abc","title":"My Project","number":16,"url":"https://github.com/users/owner/projects/16"}]}}}}`)

	runner := func(_ ...string) ([]byte, error) {
		return response, nil
	}

	cfg, err := Discover(runner, "owner", "my-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Owner != "owner" {
		t.Errorf("expected owner 'owner', got %q", cfg.Owner)
	}
	if cfg.ProjectNumber != 16 {
		t.Errorf("expected project 16, got %d", cfg.ProjectNumber)
	}
}

func TestDiscover_returnsErrorWhenNoProjects(t *testing.T) {
	t.Parallel()

	response := []byte(`{"data":{"repository":{"projectsV2":{"nodes":[]}}}}`)

	runner := func(_ ...string) ([]byte, error) {
		return response, nil
	}

	_, err := Discover(runner, "owner", "my-repo")
	if err == nil {
		t.Fatal("expected error when no projects found")
	}
}

func TestCreate_returnsNewProject(t *testing.T) {
	t.Parallel()

	response := []byte(`{"number":42,"url":"https://github.com/users/owner/projects/42","title":"my-repo"}`)

	runner := func(_ ...string) ([]byte, error) {
		return response, nil
	}

	cfg, err := Create(runner, "owner", "my-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Owner != "owner" {
		t.Errorf("expected owner 'owner', got %q", cfg.Owner)
	}
	if cfg.ProjectNumber != 42 {
		t.Errorf("expected project 42, got %d", cfg.ProjectNumber)
	}
}

func TestDiscover_returnsFirstProjectWhenMultiple(t *testing.T) {
	t.Parallel()

	response := []byte(`{"data":{"repository":{"projectsV2":{"nodes":[{"id":"PVT_1","title":"First","number":5,"url":"https://github.com/users/owner/projects/5"},{"id":"PVT_2","title":"Second","number":10,"url":"https://github.com/users/owner/projects/10"}]}}}}`)

	runner := func(_ ...string) ([]byte, error) {
		return response, nil
	}

	cfg, err := Discover(runner, "owner", "my-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ProjectNumber != 5 {
		t.Errorf("expected first project (5), got %d", cfg.ProjectNumber)
	}
}
