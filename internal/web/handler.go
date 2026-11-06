package web

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	mux       *http.ServeMux
	templates *template.Template
	webDir    string
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
	}

	staticDir := filepath.Join(webDir, "static")
	fileServer := http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))
	mux.Handle("GET /static/", fileServer)
	mux.HandleFunc("GET /{$}", s.handleIndex)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
