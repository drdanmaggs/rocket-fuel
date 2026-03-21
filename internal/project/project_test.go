package project

import (
	"strings"
	"testing"
)

func TestNextReadyReturnsFirstScoped(t *testing.T) {
	t.Parallel()

	board := &BoardSummary{
		Columns: map[string][]Item{
			"Scoped": {
				{Number: 42, Title: "First scoped issue"},
				{Number: 43, Title: "Second scoped issue"},
			},
			"In Progress": {
				{Number: 10, Title: "Already in progress"},
			},
		},
	}

	next := NextReady(board)
	if next == nil {
		t.Fatal("expected a ready issue")
	}

	if next.Number != 42 {
		t.Errorf("expected issue #42, got #%d", next.Number)
	}
}

func TestNextReadyReturnsNilWhenNoScoped(t *testing.T) {
	t.Parallel()

	board := &BoardSummary{
		Columns: map[string][]Item{
			"In Progress": {
				{Number: 10, Title: "Busy"},
			},
		},
	}

	next := NextReady(board)
	if next != nil {
		t.Errorf("expected nil, got issue #%d", next.Number)
	}
}

func TestFormatBoardOutput(t *testing.T) {
	t.Parallel()

	board := &BoardSummary{
		Columns: map[string][]Item{
			"Scoped": {
				{Number: 42, Title: "Add login", Labels: []string{"workflow:tdd"}},
			},
			"In Progress": {
				{Number: 10, Title: "Fix bug"},
			},
			"Done": {
				{Number: 1, Title: "Setup"},
				{Number: 2, Title: "CI"},
			},
		},
	}

	out := FormatBoard(board)

	checks := []string{
		"=== Project Board ===",
		"Scoped (1)",
		"#42",
		"Add login",
		"[workflow:tdd]",
		"In Progress (1)",
		"#10",
		"Done (2)",
	}

	for _, check := range checks {
		if !strings.Contains(out, check) {
			t.Errorf("expected output to contain %q\n\nGot:\n%s", check, out)
		}
	}
}

func TestFormatBoardShowsEmptyColumns(t *testing.T) {
	t.Parallel()

	board := &BoardSummary{
		Columns: map[string][]Item{},
	}

	out := FormatBoard(board)

	// All standard columns should appear with (0).
	for _, col := range []string{"Someday/Maybe", "Backlog", "Scoped", "In Progress", "Review", "Done"} {
		expected := col + " (0)"
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain %q\n\nGot:\n%s", expected, out)
		}
	}
}

func TestFormatBoardNormalizesColumnCasing(t *testing.T) {
	t.Parallel()

	board := &BoardSummary{
		Columns: map[string][]Item{
			"in progress": {
				{Number: 99, Title: "Lowercase variant"},
			},
			"IN PROGRESS": {
				{Number: 98, Title: "Uppercase variant"},
			},
			"in Review": {
				{Number: 88, Title: "Review variant"},
			},
		},
	}

	out := FormatBoard(board)

	// Both "in progress" and "IN PROGRESS" items should merge under "In Progress"
	if !strings.Contains(out, "In Progress (2)") {
		t.Errorf("expected 'In Progress (2)' in output\n\nGot:\n%s", out)
	}

	// "in Review" should merge under "In Review"
	if !strings.Contains(out, "In Review (1)") {
		t.Errorf("expected 'In Review (1)' in output\n\nGot:\n%s", out)
	}

	// Should NOT appear in catch-all section with their original casing
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "in progress (") || strings.HasPrefix(lower, "in review (") {
			// This is fine if it matches the canonical column heading exactly
			if !strings.HasPrefix(line, "In Progress (") && !strings.HasPrefix(line, "In Review (") {
				t.Errorf("case variant appeared as separate column: %q\n\nGot:\n%s", line, out)
			}
		}
	}
}

func TestFetchBoardSetsTruncatedWhenTotalCountExceedsItems(t *testing.T) {
	t.Parallel()

	fakeRun := func(args ...string) ([]byte, error) {
		// item-list returns 2 items but totalCount says 250
		if args[0] == "project" && args[1] == "item-list" {
			return []byte(`{
				"items": [
					{"id":"1","title":"Issue A","status":"Backlog","content":{"number":1,"labels":[]}},
					{"id":"2","title":"Issue B","status":"Backlog","content":{"number":2,"labels":[]}}
				],
				"totalCount": 250
			}`), nil
		}
		// project view for title
		if args[0] == "project" && args[1] == "view" {
			return []byte(`{"title":"Test Project"}`), nil
		}
		return nil, nil
	}

	board, err := FetchBoard(fakeRun, "testowner", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !board.Truncated {
		t.Error("expected Truncated to be true when totalCount > len(items)")
	}
	if board.TotalCount != 250 {
		t.Errorf("expected TotalCount=250, got %d", board.TotalCount)
	}
}

func TestFetchBoardNotTruncatedWhenAllItemsFit(t *testing.T) {
	t.Parallel()

	fakeRun := func(args ...string) ([]byte, error) {
		if args[0] == "project" && args[1] == "item-list" {
			return []byte(`{
				"items": [
					{"id":"1","title":"Issue A","status":"Backlog","content":{"number":1,"labels":[]}}
				],
				"totalCount": 1
			}`), nil
		}
		if args[0] == "project" && args[1] == "view" {
			return []byte(`{"title":"Test Project"}`), nil
		}
		return nil, nil
	}

	board, err := FetchBoard(fakeRun, "testowner", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if board.Truncated {
		t.Error("expected Truncated to be false when totalCount == len(items)")
	}
	if board.TotalCount != 1 {
		t.Errorf("expected TotalCount=1, got %d", board.TotalCount)
	}
}

func TestFormatBoardShowsTruncationWarning(t *testing.T) {
	t.Parallel()

	board := &BoardSummary{
		Columns:    map[string][]Item{},
		Truncated:  true,
		TotalCount: 250,
	}

	out := FormatBoard(board)

	if !strings.Contains(out, "WARNING") {
		t.Errorf("expected truncation warning in output\n\nGot:\n%s", out)
	}
	if !strings.Contains(out, "250") {
		t.Errorf("expected total count in warning\n\nGot:\n%s", out)
	}
}

func TestFormatBoardNoWarningWhenNotTruncated(t *testing.T) {
	t.Parallel()

	board := &BoardSummary{
		Columns:    map[string][]Item{},
		Truncated:  false,
		TotalCount: 5,
	}

	out := FormatBoard(board)

	if strings.Contains(out, "WARNING") {
		t.Errorf("expected no truncation warning\n\nGot:\n%s", out)
	}
}

func TestIsStandardColumnCaseInsensitive(t *testing.T) {
	t.Parallel()

	standard := []string{"Scoped", "In Progress", "Done"}

	if !isStandardColumn("in progress", standard) {
		t.Error("expected 'in progress' to match 'In Progress'")
	}

	if !isStandardColumn("IN PROGRESS", standard) {
		t.Error("expected 'IN PROGRESS' to match 'In Progress'")
	}

	if isStandardColumn("Custom", standard) {
		t.Error("expected 'Custom' to not be standard")
	}
}

func TestIsStandardColumn(t *testing.T) {
	t.Parallel()

	standard := []string{"Scoped", "Done"}

	if !isStandardColumn("Scoped", standard) {
		t.Error("expected 'Scoped' to be standard")
	}

	if isStandardColumn("Custom", standard) {
		t.Error("expected 'Custom' to not be standard")
	}
}
