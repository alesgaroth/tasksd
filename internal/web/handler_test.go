package web_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alesgaroth/tasksd/internal/web"
)

func TestIndexHandler(t *testing.T) {
	app := web.NewServer()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("expected Content-Type text/html, got %q", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Hello, tasksd") {
		t.Errorf("expected body to contain 'Hello, tasksd', got %q", body)
	}

	if !strings.Contains(body, "/static/htmx.min.js") {
		t.Errorf("expected body to reference '/static/htmx.min.js', got %q", body)
	}

	if !strings.Contains(body, "No items found.") {
		t.Errorf("expected empty state message, got %q", body)
	}
}

func TestIndexHandlerWithItems(t *testing.T) {
	tempDir := t.TempDir()
	itemsDir := filepath.Join(tempDir, "items")
	if err := os.MkdirAll(itemsDir, 0755); err != nil {
		t.Fatalf("failed to create items dir: %v", err)
	}

	rawItem := `---
id: T-0099
title: Web listed item
level: task
status: active
priority: A
created: 2026-11-06
---
Body text.
`
	if err := os.WriteFile(filepath.Join(itemsDir, "T-0099.md"), []byte(rawItem), 0644); err != nil {
		t.Fatalf("failed to write test item: %v", err)
	}

	app := web.NewServerWithRepo(tempDir)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "T-0099") {
		t.Errorf("expected body to contain item ID T-0099, got %q", body)
	}
	if !strings.Contains(body, "Web listed item") {
		t.Errorf("expected body to contain item title, got %q", body)
	}
}

func TestStaticAssets(t *testing.T) {
	app := web.NewServer()

	req := httptest.NewRequest(http.MethodGet, "/static/htmx.min.js", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestNotFound(t *testing.T) {
	app := web.NewServer()

	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}
