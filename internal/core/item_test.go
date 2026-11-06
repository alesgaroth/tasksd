package core_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alesgaroth/tasksd/internal/core"
)

func TestItemValidation(t *testing.T) {
	now := time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		item    core.Item
		wantErr string
	}{
		{
			name: "valid minimal task",
			item: core.Item{
				ID:      "T-0001",
				Title:   "Implement Slice 2",
				Level:   core.LevelTask,
				Status:  core.StatusActive,
				Created: now,
			},
			wantErr: "",
		},
		{
			name: "valid full task",
			item: core.Item{
				ID:        "T-0042",
				Title:     "Draft sync section 3",
				Level:     core.LevelTask,
				Parent:    "P-0007",
				Status:    core.StatusWaiting,
				Priority:  core.PriorityA,
				Urgent:    false,
				Important: true,
				MoSCoW:    core.MoSCoWShould,
				Tags:      []string{"frontier"},
				Remind:    "2026-11-12",
				WaitingOn: "Joseph",
				Links: []core.Link{
					{Type: core.LinkBlocks, To: "T-0050"},
				},
				Created: now,
				Body:    "Free-form notes.",
			},
			wantErr: "",
		},
		{
			name: "missing id",
			item: core.Item{
				Title:   "No ID",
				Level:   core.LevelTask,
				Status:  core.StatusActive,
				Created: now,
			},
			wantErr: "id is required",
		},
		{
			name: "invalid id format",
			item: core.Item{
				ID:      "invalid_id",
				Title:   "Bad ID format",
				Level:   core.LevelTask,
				Status:  core.StatusActive,
				Created: now,
			},
			wantErr: "id must match prefix-number format",
		},
		{
			name: "missing title",
			item: core.Item{
				ID:      "T-0002",
				Title:   "   ",
				Level:   core.LevelTask,
				Status:  core.StatusActive,
				Created: now,
			},
			wantErr: "title is required",
		},
		{
			name: "invalid level",
			item: core.Item{
				ID:      "T-0003",
				Title:   "Invalid level",
				Level:   core.Level("epic"),
				Status:  core.StatusActive,
				Created: now,
			},
			wantErr: "invalid level",
		},
		{
			name: "target cannot have parent",
			item: core.Item{
				ID:      "G-0001",
				Title:   "Top target",
				Level:   core.LevelTarget,
				Parent:  "P-0001",
				Status:  core.StatusActive,
				Created: now,
			},
			wantErr: "target cannot have a parent",
		},
		{
			name: "subtask must have parent",
			item: core.Item{
				ID:      "S-0001",
				Title:   "Orphan subtask",
				Level:   core.LevelSubtask,
				Status:  core.StatusActive,
				Created: now,
			},
			wantErr: "subtask must have a parent",
		},
		{
			name: "invalid status",
			item: core.Item{
				ID:      "T-0004",
				Title:   "Invalid status",
				Level:   core.LevelTask,
				Status:  core.Status("archived"),
				Created: now,
			},
			wantErr: "invalid status",
		},
		{
			name: "waiting requires waiting_on",
			item: core.Item{
				ID:      "T-0005",
				Title:   "Wait for review",
				Level:   core.LevelTask,
				Status:  core.StatusWaiting,
				Created: now,
			},
			wantErr: "waiting status requires waiting_on",
		},
		{
			name: "waiting_on prohibited when not waiting",
			item: core.Item{
				ID:        "T-0006",
				Title:     "Active item with waiting_on",
				Level:     core.LevelTask,
				Status:    core.StatusActive,
				WaitingOn: "Joseph",
				Created:   now,
			},
			wantErr: "waiting_on is only allowed when status is waiting",
		},
		{
			name: "dropped requires reason",
			item: core.Item{
				ID:      "T-0007",
				Title:   "Deprecate tool",
				Level:   core.LevelTask,
				Status:  core.StatusDropped,
				Created: now,
			},
			wantErr: "dropped status requires reason",
		},
		{
			name: "invalid priority",
			item: core.Item{
				ID:       "T-0008",
				Title:    "Bad priority",
				Level:    core.LevelTask,
				Status:   core.StatusActive,
				Priority: core.Priority("Z"),
				Created:  now,
			},
			wantErr: "invalid priority",
		},
		{
			name: "invalid moscow",
			item: core.Item{
				ID:      "T-0009",
				Title:   "Bad moscow",
				Level:   core.LevelTask,
				Status:  core.StatusActive,
				MoSCoW:  core.MoSCoW("maybe"),
				Created: now,
			},
			wantErr: "invalid moscow",
		},
		{
			name: "invalid link type",
			item: core.Item{
				ID:     "T-0010",
				Title:  "Bad link",
				Level:  core.LevelTask,
				Status: core.StatusActive,
				Links: []core.Link{
					{Type: core.LinkType("relates"), To: "T-0050"},
				},
				Created: now,
			},
			wantErr: "invalid link type",
		},
		{
			name: "link missing target",
			item: core.Item{
				ID:     "T-0011",
				Title:  "Empty link target",
				Level:  core.LevelTask,
				Status: core.StatusActive,
				Links: []core.Link{
					{Type: core.LinkBlocks, To: ""},
				},
				Created: now,
			},
			wantErr: "link destination is required",
		},
		{
			name: "missing created date",
			item: core.Item{
				ID:     "T-0012",
				Title:  "No date",
				Level:  core.LevelTask,
				Status: core.StatusActive,
			},
			wantErr: "created date is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := core.ValidateItem(tt.item)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("expected valid item, got %v", err)
			}
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
			}
		})
	}
}
