package project

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// countingRunner returns a valid board response and increments *calls on each
// item-list fetch, so tests can assert how many times the API was hit.
func countingRunner(calls *int) GHRunner {
	return func(args ...string) ([]byte, error) {
		if args[0] == "project" && args[1] == "item-list" {
			*calls++
			return []byte(`{
				"items": [
					{"id":"1","title":"Issue A","status":"Ready","content":{"number":1,"labels":["workflow:tdd"]}}
				],
				"totalCount": 1
			}`), nil
		}
		if args[0] == "project" && args[1] == "view" {
			return []byte(`{"title":"Test Project"}`), nil
		}
		return nil, nil
	}
}

func TestFetchBoardCachedFetchesFreshOnCacheMiss(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	calls := 0
	now := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)

	board, err := FetchBoardCached(countingRunner(&calls), "owner", 1, repoDir, 60*time.Second, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 fetch on cache miss, got %d", calls)
	}
	if board == nil || len(board.Columns["Ready"]) != 1 {
		t.Fatal("expected board with one Ready item")
	}

	// The cache file should now exist.
	cachePath := filepath.Join(repoDir, ".rocket-fuel", "board-cache.json")
	if _, statErr := os.Stat(cachePath); statErr != nil {
		t.Errorf("expected cache file to be written: %v", statErr)
	}
}

func TestFetchBoardCachedServesFromCacheWithinTTL(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	calls := 0
	now := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)
	run := countingRunner(&calls)

	// First call populates the cache.
	if _, err := FetchBoardCached(run, "owner", 1, repoDir, 60*time.Second, now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Second call 30s later (within the 60s TTL) must hit the cache.
	board, err := FetchBoardCached(run, "owner", 1, repoDir, 60*time.Second, now.Add(30*time.Second))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if calls != 1 {
		t.Errorf("expected cache hit (1 fetch total), got %d fetches", calls)
	}
	if board == nil || len(board.Columns["Ready"]) != 1 {
		t.Fatal("expected cached board to round-trip the Ready item")
	}
}

func TestFetchBoardCachedRefetchesAfterTTLExpiry(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	calls := 0
	now := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)
	run := countingRunner(&calls)

	if _, err := FetchBoardCached(run, "owner", 1, repoDir, 60*time.Second, now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 61s later — past the 60s TTL, so it must fetch fresh.
	if _, err := FetchBoardCached(run, "owner", 1, repoDir, 60*time.Second, now.Add(61*time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if calls != 2 {
		t.Errorf("expected 2 fetches after TTL expiry, got %d", calls)
	}
}

func TestFetchBoardCachedRefetchesWhenCacheCorrupt(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	cacheDir := filepath.Join(repoDir, ".rocket-fuel")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "board-cache.json"), []byte("{not valid json"), 0o644); err != nil {
		t.Fatal(err)
	}

	calls := 0
	now := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)

	board, err := FetchBoardCached(countingRunner(&calls), "owner", 1, repoDir, 60*time.Second, now)
	if err != nil {
		t.Fatalf("expected corrupt cache to be ignored, got error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected fresh fetch when cache is corrupt, got %d fetches", calls)
	}
	if board == nil {
		t.Fatal("expected a board despite corrupt cache")
	}
}

func TestFetchBoardCachedPropagatesFetchError(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	now := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)

	failing := func(_ ...string) ([]byte, error) {
		return nil, os.ErrPermission
	}

	if _, err := FetchBoardCached(failing, "owner", 1, repoDir, 60*time.Second, now); err == nil {
		t.Error("expected error to propagate when the underlying fetch fails")
	}
}
