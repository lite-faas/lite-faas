package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/litefaas/litefaas/internal/container"
	"github.com/litefaas/litefaas/internal/database"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	containerMgr *container.Manager
	db           *database.Database
	logger       *logrus.Logger
}

func NewHandler(containerMgr *container.Manager, db *database.Database, logger *logrus.Logger) *Handler {
	return &Handler{
		containerMgr: containerMgr,
		db:           db,
		logger:       logger,
	}
}

func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		http.Redirect(w, r, "/web/", http.StatusFound)
		return
	}

	if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/web/") {
		http.NotFound(w, r)
		return
	}

	functionName := strings.TrimPrefix(path, "/")
	if functionName == "" {
		http.NotFound(w, r)
		return
	}

	function, err := h.db.GetFunction(functionName)
	if err != nil {
		h.logger.Errorf("Function not found: %s", functionName)
		http.NotFound(w, r)
		return
	}

	if function.Status != "running" {
		h.logger.Errorf("Function %s is not running (status: %s)", functionName, function.Status)
		http.Error(w, "Function not available", http.StatusServiceUnavailable)
		return
	}

	targetURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("127.0.0.1:%d", function.Port),
		Path:   r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(w, r)
}
