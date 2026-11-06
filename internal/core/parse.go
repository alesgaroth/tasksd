package core

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type frontmatter struct {
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
	Created   string    `yaml:"created"`
}

func ParseItem(raw string) (Item, error) {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	trimmed := strings.TrimSpace(normalized)
	if !strings.HasPrefix(trimmed, "---") {
		return Item{}, errors.New("missing frontmatter delimiter at start")
	}

	rest := trimmed[3:]
	if strings.HasPrefix(rest, "\n") {
		rest = rest[1:]
	}

	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		return Item{}, errors.New("missing frontmatter closing delimiter")
	}

	yamlBlock := rest[:endIdx]
	bodyBlock := rest[endIdx+4:]
	if strings.HasPrefix(bodyBlock, "\n") {
		bodyBlock = bodyBlock[1:]
	}

	var fm frontmatter
	decoder := yaml.NewDecoder(strings.NewReader(yamlBlock))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fm); err != nil {
		return Item{}, fmt.Errorf("yaml parse error: %w", err)
	}

	var createdTime time.Time
	if fm.Created != "" {
		t, err := time.Parse("2006-01-02", strings.Trim(fm.Created, `"`))
		if err != nil {
			return Item{}, fmt.Errorf("invalid created date format: %w", err)
		}
		createdTime = t
	}

	item := Item{
		ID:        fm.ID,
		Title:     fm.Title,
		Level:     fm.Level,
		Parent:    fm.Parent,
		Status:    fm.Status,
		Priority:  fm.Priority,
		Urgent:    fm.Urgent,
		Important: fm.Important,
		MoSCoW:    fm.MoSCoW,
		Tags:      fm.Tags,
		Remind:    fm.Remind,
		WaitingOn: fm.WaitingOn,
		Reason:    fm.Reason,
		Links:     fm.Links,
		Created:   createdTime,
		Body:      bodyBlock,
	}

	if err := ValidateItem(item); err != nil {
		return Item{}, err
	}

	return item, nil
}

func FormatItem(item Item) (string, error) {
	if err := ValidateItem(item); err != nil {
		return "", fmt.Errorf("cannot format invalid item: %w", err)
	}

	fm := frontmatter{
		ID:        item.ID,
		Title:     item.Title,
		Level:     item.Level,
		Parent:    item.Parent,
		Status:    item.Status,
		Priority:  item.Priority,
		Urgent:    item.Urgent,
		Important: item.Important,
		MoSCoW:    item.MoSCoW,
		Tags:      item.Tags,
		Remind:    item.Remind,
		WaitingOn: item.WaitingOn,
		Reason:    item.Reason,
		Links:     item.Links,
		Created:   item.Created.Format("2006-01-02"),
	}

	yamlBytes, _ := yaml.Marshal(fm)

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(yamlBytes)
	buf.WriteString("---\n")
	if item.Body != "" {
		buf.WriteString(item.Body)
		if !strings.HasSuffix(item.Body, "\n") {
			buf.WriteString("\n")
		}
	}

	return buf.String(), nil
}
