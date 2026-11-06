package store_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alesgaroth/tasksd/internal/core"
	"github.com/alesgaroth/tasksd/internal/store"
)

func TestStoreListItems(t *testing.T) {
	tempDir := t.TempDir()
	itemsDir := filepath.Join(tempDir, "items")
	if err := os.MkdirAll(itemsDir, 0755); err != nil {
		t.Fatalf("failed to create items dir: %v", err)
	}

	validItem := `---
id: T-0001
title: First test item
level: task
status: active
created: 2026-11-06
---
Some notes.
`
	if err := os.WriteFile(filepath.Join(itemsDir, "T-0001.md"), []byte(validItem), 0644); err != nil {
		t.Fatalf("failed to write test item: %v", err)
	}

	// Non-markdown file to ensure it's ignored
	if err := os.WriteFile(filepath.Join(itemsDir, "notes.txt"), []byte("ignore me"), 0644); err != nil {
		t.Fatalf("failed to write non-md file: %v", err)
	}

	s := store.New(tempDir)
	items, err := s.ListItems()
	if err != nil {
		t.Fatalf("unexpected error listing items: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "T-0001" || items[0].Title != "First test item" || items[0].Level != core.LevelTask {
		t.Errorf("unexpected item content: %+v", items[0])
	}
}

func TestStoreNonExistentDir(t *testing.T) {
	s := store.New("/non/existent/path")
	items, err := s.ListItems()
	if err != nil {
		t.Fatalf("expected no error for empty/non-existent dirs, got %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestStoreInvalidItemError(t *testing.T) {
	tempDir := t.TempDir()
	itemsDir := filepath.Join(tempDir, "items")
	if err := os.MkdirAll(itemsDir, 0755); err != nil {
		t.Fatalf("failed to create items dir: %v", err)
	}

	invalidContent := "not-frontmatter"
	if err := os.WriteFile(filepath.Join(itemsDir, "bad.md"), []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to write bad item: %v", err)
	}

	s := store.New(tempDir)
	_, err := s.ListItems()
	if err == nil {
		t.Fatalf("expected parse error, got nil")
	}
}

func TestValidatePath(t *testing.T) {
	tempDir := t.TempDir()
	itemsDir := filepath.Join(tempDir, "items")
	if err := os.MkdirAll(itemsDir, 0755); err != nil {
		t.Fatalf("failed to create items dir: %v", err)
	}

	validItem := `---
id: T-0001
title: Valid item
level: task
status: active
created: 2026-11-06
---
Notes.
`
	validPath := filepath.Join(itemsDir, "T-0001.md")
	if err := os.WriteFile(validPath, []byte(validItem), 0644); err != nil {
		t.Fatalf("failed to write valid item: %v", err)
	}

	// 1. Single valid file
	errs, count, err := store.ValidatePath(validPath)
	if err != nil || len(errs) != 0 || count != 1 {
		t.Fatalf("expected 1 valid item with 0 errors, got count=%d, errs=%v, err=%v", count, errs, err)
	}

	// 2. Add an invalid file
	invalidItem := `---
id: T-0002
title: Missing created
level: task
status: active
---
`
	invalidPath := filepath.Join(itemsDir, "T-0002.md")
	if err := os.WriteFile(invalidPath, []byte(invalidItem), 0644); err != nil {
		t.Fatalf("failed to write invalid item: %v", err)
	}

	// Single invalid file
	errs, count, err = store.ValidatePath(invalidPath)
	if err != nil || len(errs) != 1 || count != 1 {
		t.Fatalf("expected 1 item with 1 error, got count=%d, errs=%v, err=%v", count, errs, err)
	}

	// Directory validation
	errs, count, err = store.ValidatePath(tempDir)
	if err != nil || len(errs) != 1 || count != 2 {
		t.Fatalf("expected 2 items with 1 error, got count=%d, errs=%v, err=%v", count, errs, err)
	}

	// Non-existent path
	_, _, err = store.ValidatePath(filepath.Join(tempDir, "nonexistent"))
	if err == nil {
		t.Fatalf("expected error for non-existent path")
	}
}

func TestStoreCaptureItem(t *testing.T) {
	tempDir := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %s: %v", string(out), err)
	}

	s := store.New(tempDir)
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	item := core.Item{
		ID:      "T-0001",
		Title:   "First captured task",
		Level:   core.LevelTask,
		Status:  core.StatusInbox,
		Created: now,
		Body:    "Some note",
	}

	if err := s.CaptureItem(item); err != nil {
		t.Fatalf("failed to capture item: %v", err)
	}

	filePath := filepath.Join(tempDir, "items", "T-0001.md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read created item: %v", err)
	}

	parsed, err := core.ParseItem(string(content))
	if err != nil {
		t.Fatalf("failed to parse created item: %v", err)
	}
	if parsed.ID != item.ID || parsed.Title != item.Title {
		t.Errorf("parsed item mismatch: got %+v, want %+v", parsed, item)
	}

	logCmd := exec.Command("git", "log", "-1", "--pretty=%B")
	logCmd.Dir = tempDir
	logOut, err := logCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log failed: %s: %v", string(logOut), err)
	}
	if !strings.Contains(string(logOut), "Capture T-0001: First captured task") {
		t.Errorf("unexpected commit message: %s", string(logOut))
	}

	nextID, err := s.NextID("T")
	if err != nil {
		t.Fatalf("failed to get next ID: %v", err)
	}
	if nextID != "T-0002" {
		t.Errorf("expected next ID T-0002, got %q", nextID)
	}
}
