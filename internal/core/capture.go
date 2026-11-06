package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type CaptureParams struct {
	ID    string
	Title string
	Body  string
	Today time.Time
}

func Capture(params CaptureParams) (Item, error) {
	item := Item{
		ID:      params.ID,
		Title:   strings.TrimSpace(params.Title),
		Level:   LevelTask,
		Status:  StatusInbox,
		Created: params.Today,
		Body:    strings.TrimSpace(params.Body),
	}
	if err := ValidateItem(item); err != nil {
		return Item{}, err
	}
	return item, nil
}

func NextID(prefix string, existingIDs []string) string {
	maxNum := 0
	prefixWithDash := prefix + "-"
	for _, id := range existingIDs {
		if strings.HasPrefix(id, prefixWithDash) {
			suffix := strings.TrimPrefix(id, prefixWithDash)
			if num, err := strconv.Atoi(suffix); err == nil {
				if num > maxNum {
					maxNum = num
				}
			}
		}
	}
	return fmt.Sprintf("%s-%04d", prefix, maxNum+1)
}
