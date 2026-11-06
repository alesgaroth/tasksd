package core

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var idRegex = regexp.MustCompile(`^[A-Za-z]+-[0-9]+$`)

type Level string

const (
	LevelTarget  Level = "target"
	LevelProject Level = "project"
	LevelTask    Level = "task"
	LevelSubtask Level = "subtask"
)

type Status string

const (
	StatusInbox   Status = "inbox"
	StatusActive  Status = "active"
	StatusWaiting Status = "waiting"
	StatusDone    Status = "done"
	StatusDropped Status = "dropped"
)

type Priority string

const (
	PriorityA Priority = "A"
	PriorityB Priority = "B"
	PriorityC Priority = "C"
	PriorityD Priority = "D"
	PriorityE Priority = "E"
)

type MoSCoW string

const (
	MoSCoWMust   MoSCoW = "must"
	MoSCoWShould MoSCoW = "should"
	MoSCoWCould  MoSCoW = "could"
	MoSCoWWont   MoSCoW = "wont"
)

type LinkType string

const (
	LinkBlocks    LinkType = "blocks"
	LinkRelated   LinkType = "related"
	LinkDuplicate LinkType = "duplicate"
)

type Link struct {
	Type LinkType `yaml:"type"`
	To   string   `yaml:"to"`
}

type Item struct {
	ID        string    `yaml:"id"`
	Title     string    `yaml:"title"`
	Level     Level     `yaml:"level"`
	Parent    string    `yaml:"parent,omitempty"`
	Status    Status    `yaml:"status"`
	Priority  Priority  `yaml:"priority,omitempty"`
	Urgent    bool      `yaml:"urgent"`
	Important bool      `yaml:"important"`
	MoSCoW    MoSCoW    `yaml:"moscow,omitempty"`
	Tags      []string  `yaml:"tags,omitempty"`
	Remind    string    `yaml:"remind,omitempty"`
	WaitingOn string    `yaml:"waiting_on,omitempty"`
	Reason    string    `yaml:"reason,omitempty"`
	Links     []Link    `yaml:"links,omitempty"`
	Created   time.Time `yaml:"created"`
	Body      string    `yaml:"-"`
}

func ValidateItem(item Item) error {
	if item.ID == "" {
		return errors.New("id is required")
	}
	if !idRegex.MatchString(item.ID) {
		return errors.New("id must match prefix-number format")
	}

	if strings.TrimSpace(item.Title) == "" {
		return errors.New("title is required")
	}

	switch item.Level {
	case LevelTarget, LevelProject, LevelTask, LevelSubtask:
	default:
		return fmt.Errorf("invalid level: %q", item.Level)
	}

	if item.Level == LevelTarget && item.Parent != "" {
		return errors.New("target cannot have a parent")
	}
	if item.Level == LevelSubtask && item.Parent == "" {
		return errors.New("subtask must have a parent")
	}

	switch item.Status {
	case StatusInbox, StatusActive, StatusWaiting, StatusDone, StatusDropped:
	default:
		return fmt.Errorf("invalid status: %q", item.Status)
	}

	if item.Status == StatusWaiting && strings.TrimSpace(item.WaitingOn) == "" {
		return errors.New("waiting status requires waiting_on")
	}
	if item.Status != StatusWaiting && strings.TrimSpace(item.WaitingOn) != "" {
		return errors.New("waiting_on is only allowed when status is waiting")
	}

	if item.Status == StatusDropped && strings.TrimSpace(item.Reason) == "" {
		return errors.New("dropped status requires reason")
	}

	if item.Priority != "" {
		switch item.Priority {
		case PriorityA, PriorityB, PriorityC, PriorityD, PriorityE:
		default:
			return fmt.Errorf("invalid priority: %q", item.Priority)
		}
	}

	if item.MoSCoW != "" {
		switch item.MoSCoW {
		case MoSCoWMust, MoSCoWShould, MoSCoWCould, MoSCoWWont:
		default:
			return fmt.Errorf("invalid moscow: %q", item.MoSCoW)
		}
	}

	for _, link := range item.Links {
		switch link.Type {
		case LinkBlocks, LinkRelated, LinkDuplicate:
		default:
			return fmt.Errorf("invalid link type: %q", link.Type)
		}
		if strings.TrimSpace(link.To) == "" {
			return errors.New("link destination is required")
		}
	}

	if item.Created.IsZero() {
		return errors.New("created date is required")
	}

	return nil
}
