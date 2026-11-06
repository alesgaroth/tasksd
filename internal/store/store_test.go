package store_test

import (
	"os"
	"path/filepath"
	"testing"

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
