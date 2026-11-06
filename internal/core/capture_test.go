package core_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alesgaroth/tasksd/internal/core"
)

func TestCapture(t *testing.T) {
	today := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)

	item, err := core.Capture(core.CaptureParams{
		ID:    "T-0001",
		Title: "Test capture task",
		Body:  "Context notes",
		Today: today,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if item.ID != "T-0001" {
		t.Errorf("expected ID T-0001, got %q", item.ID)
	}
	if item.Level != core.LevelTask {
		t.Errorf("expected level task, got %q", item.Level)
	}
	if item.Status != core.StatusInbox {
		t.Errorf("expected status inbox, got %q", item.Status)
	}
	if !item.Created.Equal(today) {
		t.Errorf("expected created date %v, got %v", today, item.Created)
	}
	if item.Body != "Context notes" {
		t.Errorf("expected body 'Context notes', got %q", item.Body)
	}

	// Missing title validation
	_, err = core.Capture(core.CaptureParams{
		ID:    "T-0002",
		Title: "   ",
		Today: today,
	})
	if err == nil || !strings.Contains(err.Error(), "title is required") {
		t.Fatalf("expected title required error, got %v", err)
	}
}

func TestNextID(t *testing.T) {
	tests := []struct {
		prefix      string
		existingIDs []string
		want        string
	}{
		{prefix: "T", existingIDs: nil, want: "T-0001"},
		{prefix: "T", existingIDs: []string{"T-0001", "T-0002"}, want: "T-0003"},
		{prefix: "T", existingIDs: []string{"T-0001", "P-0009", "T-0042"}, want: "T-0043"},
		{prefix: "P", existingIDs: []string{"P-0007", "T-0010"}, want: "P-0008"},
		{prefix: "T", existingIDs: []string{"T-invalid", "X-100"}, want: "T-0001"},
	}

	for _, tt := range tests {
		got := core.NextID(tt.prefix, tt.existingIDs)
		if got != tt.want {
			t.Errorf("NextID(%q, %v) = %q, want %q", tt.prefix, tt.existingIDs, got, tt.want)
		}
	}
}
