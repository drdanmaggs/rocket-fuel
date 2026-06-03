package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// boardCacheFile is the on-disk cache location relative to the repo root.
const boardCacheFile = "board-cache.json"

// Board cache TTLs for the hook-driven callers (#213).
const (
	// StopHookBoardTTL caches the board for the Stop hook, which fires after
	// every Claude response and is the worst rate-limit offender.
	StopHookBoardTTL = 60 * time.Second
	// SessionBoardTTL caches the board for the session-priming hooks (prime,
	// precompact), which fire far less often than the Stop hook.
	SessionBoardTTL = 30 * time.Second
)

// boardCacheEnvelope wraps a cached board with the time it was fetched.
type boardCacheEnvelope struct {
	FetchedAt time.Time     `json:"fetched_at"`
	Board     *BoardSummary `json:"board"`
}

// FetchBoardCached returns the project board, serving from a local cache file
// when the cached copy is younger than ttl and otherwise fetching fresh via
// run and refreshing the cache.
//
// Hooks that fire on every Claude response (the Stop hook in particular) call
// this to avoid exhausting the GitHub API rate limit — see issue #213. now is
// injected so the TTL behaviour is deterministic in tests.
//
// Caching is a best-effort optimisation: a corrupt or unwritable cache never
// fails the call, it just falls through to a fresh fetch.
func FetchBoardCached(run GHRunner, owner string, projectNumber int, repoDir string, ttl time.Duration, now time.Time) (*BoardSummary, error) {
	cachePath := filepath.Join(repoDir, ".rocket-fuel", boardCacheFile)

	if cached := readBoardCache(cachePath); cached != nil && cached.Board != nil {
		if now.Sub(cached.FetchedAt) < ttl {
			return cached.Board, nil
		}
	}

	board, err := FetchBoard(run, owner, projectNumber)
	if err != nil {
		return nil, err
	}

	writeBoardCache(cachePath, boardCacheEnvelope{FetchedAt: now, Board: board})
	return board, nil
}

// readBoardCache loads the cache envelope, returning nil on any error so the
// caller falls through to a fresh fetch.
func readBoardCache(cachePath string) *boardCacheEnvelope {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil
	}
	var env boardCacheEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil
	}
	return &env
}

// writeBoardCache persists the envelope, ignoring errors — caching must never
// break the calling hook.
func writeBoardCache(cachePath string, env boardCacheEnvelope) {
	data, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return
	}
	_ = os.WriteFile(cachePath, data, 0o644)
}
