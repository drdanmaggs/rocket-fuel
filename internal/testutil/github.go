package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// FakeGitHub creates a test HTTP server that responds to GitHub API requests.
// Returns the server (caller should defer server.Close()) and its URL.
func FakeGitHub(t *testing.T, responses map[string]any) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, ok := responses[r.Method+" "+r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response for %s %s: %v", r.Method, r.URL.Path, err)
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}))

	return server
}
