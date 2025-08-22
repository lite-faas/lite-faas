package web

import (
	"embed"
	"html/template"
	"net/http"
	"strings"
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

	if strings.HasPrefix(r.URL.Path, "/static/") {
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

	// Définir le bon type MIME selon l'extension
	switch {
	case strings.HasSuffix(path, ".js"):
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case strings.HasSuffix(path, ".css"):
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case strings.HasSuffix(path, ".html"):
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case strings.HasSuffix(path, ".json"):
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	case strings.HasSuffix(path, ".png"):
		w.Header().Set("Content-Type", "image/png")
	case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
		w.Header().Set("Content-Type", "image/jpeg")
	case strings.HasSuffix(path, ".gif"):
		w.Header().Set("Content-Type", "image/gif")
	case strings.HasSuffix(path, ".svg"):
		w.Header().Set("Content-Type", "image/svg+xml")
	case strings.HasSuffix(path, ".ico"):
		w.Header().Set("Content-Type", "image/x-icon")
	default:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}

	// Ajouter des headers de cache pour les fichiers statiques
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	w.Write(content)
}
