package web_test

import (
	"net/http"
	"net/http/httptest"
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
