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
