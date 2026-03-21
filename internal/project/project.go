// Package project provides GitHub Project board integration for the Integrator.
package project

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Config stores the linked GitHub Project reference.
type Config struct {
	ProjectNumber int    `json:"project_number"`
	Owner         string `json:"owner"`
}

// Item represents an issue on the project board.
type Item struct {
	ID     string
	Title  string
	Number int
	Status string
	Labels []string
}

// BoardSummary holds the state of the project board by column.
type BoardSummary struct {
	ProjectTitle string
	Columns      map[string][]Item
}

// FetchBoard reads the current state of a GitHub Project board.
func FetchBoard(run GHRunner, owner string, projectNumber int) (*BoardSummary, error) {
	out, err := run("project", "item-list", strconv.Itoa(projectNumber),
		"--owner", owner,
		"--format", "json",
		"--limit", "200",
	)
	if err != nil {
		return nil, fmt.Errorf("gh project item-list: %w", err)
	}

	var raw struct {
		Items []struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			Status  string `json:"status"`
			Content struct {
				Number int      `json:"number"`
				Labels []string `json:"labels"`
			} `json:"content"`
		} `json:"items"`
		TotalCount int `json:"totalCount"`
	}

	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse project items: %w", err)
	}

	summary := &BoardSummary{
		Columns: make(map[string][]Item),
	}

	for _, item := range raw.Items {
		status := item.Status
		if status == "" {
			status = "No Status"
		}

		summary.Columns[status] = append(summary.Columns[status], Item{
			ID:     item.ID,
			Title:  item.Title,
			Number: item.Content.Number,
			Status: status,
			Labels: item.Content.Labels,
		})
	}

	return summary, nil
}

// readyColumnNames are the column names that indicate work is ready to dispatch.
// GitHub's default Kanban uses "Ready". Rocket Fuel docs use "Scoped".
var readyColumnNames = []string{"Ready", "Scoped"}

// NextReady returns the highest-priority issue from the ready column.
// Checks for "Ready" (GitHub default) and "Scoped" (Rocket Fuel convention).
// Returns nil if no ready issues are available.
func NextReady(board *BoardSummary) *Item {
	for _, name := range readyColumnNames {
		items := board.Columns[name]
		if len(items) > 0 {
			return &items[0]
		}
	}
	return nil
}

// FormatBoard renders the board as a human-readable summary.
func FormatBoard(board *BoardSummary) string {
	var b strings.Builder

	columnOrder := []string{"Someday/Maybe", "Backlog", "Ready", "Scoped", "In Progress", "In progress", "Review", "In review", "Done"}

	_, _ = fmt.Fprintln(&b, "=== Project Board ===")
	_, _ = fmt.Fprintln(&b)

	for _, col := range columnOrder {
		items := board.Columns[col]
		_, _ = fmt.Fprintf(&b, "%s (%d)\n", col, len(items))

		for _, item := range items {
			labels := ""
			if len(item.Labels) > 0 {
				labels = " [" + strings.Join(item.Labels, ", ") + "]"
			}
			_, _ = fmt.Fprintf(&b, "  #%-4d %s%s\n", item.Number, item.Title, labels)
		}

		_, _ = fmt.Fprintln(&b)
	}

	// Any columns not in the standard order.
	for col, items := range board.Columns {
		if isStandardColumn(col, columnOrder) {
			continue
		}
		_, _ = fmt.Fprintf(&b, "%s (%d)\n", col, len(items))
		for _, item := range items {
			_, _ = fmt.Fprintf(&b, "  #%-4d %s\n", item.Number, item.Title)
		}
		_, _ = fmt.Fprintln(&b)
	}

	return b.String()
}

func isStandardColumn(col string, standard []string) bool {
	for _, s := range standard {
		if col == s {
			return true
		}
	}
	return false
}
