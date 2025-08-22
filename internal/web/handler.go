package web

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed templates/*
var templates embed.FS

//go:embed static/*
var staticFiles embed.FS

type Handler struct {
	templates *template.Template
}

func NewHandler() *Handler {
	tmpl, err := template.ParseFS(templates, "templates/*.html")
	if err != nil {
		panic(err)
	}

	return &Handler{
		templates: tmpl,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		h.serveDashboard(w, r)
		return
	}

	if r.URL.Path == "/static/" {
		h.serveStatic(w, r)
		return
	}

	http.NotFound(w, r)
}

func (h *Handler) serveDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	h.templates.ExecuteTemplate(w, "dashboard.html", nil)
}

func (h *Handler) serveStatic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/static/"):]
	if path == "" {
		http.NotFound(w, r)
		return
	}

	content, err := staticFiles.ReadFile("static/" + path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch {
	case len(path) > 3 && path[len(path)-3:] == ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case len(path) > 4 && path[len(path)-4:] == ".css":
		w.Header().Set("Content-Type", "text/css")
	default:
		w.Header().Set("Content-Type", "text/plain")
	}

	w.Write(content)
}
