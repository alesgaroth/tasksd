package web_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alesgaroth/tasksd/internal/web"
)

func initTempGitRepo(t *testing.T) string {
	tempDir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %s: %v", string(out), err)
	}
	return tempDir
}

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
	if !strings.Contains(body, "tasksd") {
		t.Errorf("expected body to contain 'tasksd', got %q", body)
	}

	if !strings.Contains(body, "/static/htmx.min.js") {
		t.Errorf("expected body to reference '/static/htmx.min.js', got %q", body)
	}

	if !strings.Contains(body, `action=""`) && !strings.Contains(body, `hx-post="/capture"`) {
		t.Errorf("expected capture form in body, got %q", body)
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

func TestCaptureHTMX(t *testing.T) {
	tempDir := initTempGitRepo(t)
	app := web.NewServerWithRepo(tempDir)

	form := url.Values{}
	form.Set("title", "Captured HTMX Task")
	form.Set("body", "HTMX context notes")

	req := httptest.NewRequest(http.MethodPost, "/capture", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "T-0001") {
		t.Errorf("expected body to contain T-0001, got %q", body)
	}
	if !strings.Contains(body, "Captured HTMX Task") {
		t.Errorf("expected body to contain task title, got %q", body)
	}

	// Verify file on disk
	filePath := filepath.Join(tempDir, "items", "T-0001.md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("item file not created: %v", err)
	}
	if !strings.Contains(string(content), "Captured HTMX Task") {
		t.Errorf("item file missing title: %s", string(content))
	}

	// Verify git commit
	logCmd := exec.Command("git", "log", "-1", "--pretty=%B")
	logCmd.Dir = tempDir
	logOut, err := logCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log failed: %s: %v", string(logOut), err)
	}
	if !strings.Contains(string(logOut), "Capture T-0001: Captured HTMX Task") {
		t.Errorf("unexpected commit message: %s", string(logOut))
	}
}

func TestCaptureStandardRedirect(t *testing.T) {
	tempDir := initTempGitRepo(t)
	app := web.NewServerWithRepo(tempDir)

	form := url.Values{}
	form.Set("title", "Redirect Task")

	req := httptest.NewRequest(http.MethodPost, "/capture", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", rec.Code)
	}

	if loc := rec.Header().Get("Location"); loc != "/" {
		t.Errorf("expected redirect to '/', got %q", loc)
	}
}

func TestCaptureValidationFailure(t *testing.T) {
	tempDir := initTempGitRepo(t)
	app := web.NewServerWithRepo(tempDir)

	form := url.Values{}
	form.Set("title", "   ")

	req := httptest.NewRequest(http.MethodPost, "/capture", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for empty title, got %d", rec.Code)
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
