package web

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/alesgaroth/tasksd/internal/core"
	"github.com/alesgaroth/tasksd/internal/store"
)

type Server struct {
	mux       *http.ServeMux
	templates *template.Template
	webDir    string
	store     *store.Store
}

type PageData struct {
	Items []core.Item
}

func findWebDir() string {
	candidates := []string{"web", "../../web", "../web"}
	for _, dir := range candidates {
		tmplPath := filepath.Join(dir, "templates", "index.html")
		if _, err := os.Stat(tmplPath); err == nil {
			return dir
		}
	}
	return "web"
}

func NewServer() http.Handler {
	return NewServerWithRepo(".")
}

func NewServerWithRepo(repoDir string) http.Handler {
	webDir := findWebDir()
	tmplPath := filepath.Join(webDir, "templates", "index.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	s := &Server{
		mux:       mux,
		templates: tmpl,
		webDir:    webDir,
		store:     store.New(repoDir),
	}

	staticDir := filepath.Join(webDir, "static")
	fileServer := http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))
	mux.Handle("GET /static/", fileServer)
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("POST /capture", s.handleCapture)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	items, err := s.store.ListItems()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := PageData{
		Items: items,
	}
	if err := s.templates.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleCapture(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	body := r.FormValue("body")

	nextID, err := s.store.NextID("T")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	item, err := core.Capture(core.CaptureParams{
		ID:    nextID,
		Title: title,
		Body:  body,
		Today: time.Now(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.store.CaptureItem(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		rowHTML := fmt.Sprintf(`<tr id="item-%s"><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
			template.HTMLEscapeString(item.ID),
			template.HTMLEscapeString(item.ID),
			template.HTMLEscapeString(item.Title),
			template.HTMLEscapeString(string(item.Level)),
			template.HTMLEscapeString(string(item.Status)),
			template.HTMLEscapeString(string(item.Priority)),
		)
		w.Write([]byte(rowHTML))
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
