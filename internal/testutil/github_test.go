package testutil

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestFakeGitHubServesResponses(t *testing.T) {
	t.Parallel()

	server := FakeGitHub(t, map[string]any{
		"GET /repos/test/repo/issues": []map[string]any{
			{"number": 1, "title": "Test issue"},
		},
	})
	defer server.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/repos/test/repo/issues", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	var issues []map[string]any
	if err := json.Unmarshal(body, &issues); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}

	if issues[0]["title"] != "Test issue" {
		t.Errorf("expected title 'Test issue', got %v", issues[0]["title"])
	}
}

func TestFakeGitHubReturns404ForUnknownRoutes(t *testing.T) {
	t.Parallel()

	server := FakeGitHub(t, map[string]any{})
	defer server.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/unknown", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
