package core_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alesgaroth/tasksd/internal/core"
)

func TestParseItemRoundTrip(t *testing.T) {
	raw := `---
id: T-0042
title: Draft sync section 3
level: task
parent: P-0007
status: waiting
priority: A
urgent: false
important: true
moscow: should
tags:
    - frontier
remind: "2026-11-12"
waiting_on: Joseph
links:
    - type: blocks
      to: T-0050
created: 2026-11-05
---
Free-form notes.
`

	item, err := core.ParseItem(raw)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if item.ID != "T-0042" {
		t.Errorf("expected ID T-0042, got %q", item.ID)
	}
	if item.Title != "Draft sync section 3" {
		t.Errorf("expected Title 'Draft sync section 3', got %q", item.Title)
	}
	if item.Level != core.LevelTask {
		t.Errorf("expected Level task, got %q", item.Level)
	}
	if item.Parent != "P-0007" {
		t.Errorf("expected Parent P-0007, got %q", item.Parent)
	}
	if item.Status != core.StatusWaiting {
		t.Errorf("expected Status waiting, got %q", item.Status)
	}
	if item.WaitingOn != "Joseph" {
		t.Errorf("expected WaitingOn Joseph, got %q", item.WaitingOn)
	}
	if len(item.Links) != 1 || item.Links[0].Type != core.LinkBlocks || item.Links[0].To != "T-0050" {
		t.Errorf("unexpected links: %+v", item.Links)
	}
	if strings.TrimSpace(item.Body) != "Free-form notes." {
		t.Errorf("unexpected body: %q", item.Body)
	}

	formatted, err := core.FormatItem(item)
	if err != nil {
		t.Fatalf("unexpected format error: %v", err)
	}

	roundTripped, err := core.ParseItem(formatted)
	if err != nil {
		t.Fatalf("unexpected re-parse error: %v", err)
	}

	if roundTripped.ID != item.ID || roundTripped.Title != item.Title || roundTripped.Body != item.Body {
		t.Errorf("round trip mismatch: got %+v, want %+v", roundTripped, item)
	}
}

func TestParseItemErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "empty file",
			input:   "",
			wantErr: "missing frontmatter",
		},
		{
			name:    "no frontmatter delimiters",
			input:   "Just notes without frontmatter",
			wantErr: "missing frontmatter delimiter at start",
		},
		{
			name: "missing closing delimiter",
			input: `---
id: T-0001
title: No closing delimiter`,
			wantErr: "missing frontmatter closing delimiter",
		},
		{
			name: "invalid created date",
			input: `---
id: T-0001
title: Bad date
level: task
status: active
created: "not-a-date"
---
Body`,
			wantErr: "invalid created date format",
		},
		{
			name: "invalid yaml",
			input: `---
id: [invalid: yaml
---
Body`,
			wantErr: "yaml parse error",
		},
		{
			name: "validation failure",
			input: `---
id: T-0001
title: Task without created
level: task
status: active
---
Notes`,
			wantErr: "created date is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := core.ParseItem(tt.input)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestFormatItemValidationError(t *testing.T) {
	invalidItem := core.Item{
		ID:      "invalid",
		Title:   "Invalid item",
		Created: time.Now(),
	}
	_, err := core.FormatItem(invalidItem)
	if err == nil {
		t.Fatalf("expected error formatting invalid item, got nil")
	}
}
