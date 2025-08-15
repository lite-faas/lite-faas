package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/litefaas/litefaas/internal/container"
	"github.com/litefaas/litefaas/internal/database"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	db           *database.Database
	containerMgr *container.Manager
	logger       *logrus.Logger
}

type CreateFunctionRequest struct {
	Name     string `json:"name"`
	Language string `json:"language"`
	Code     string `json:"code"`
}

type FunctionResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Language  string `json:"language"`
	Code      string `json:"code"`
	Port      int    `json:"port"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func NewHandler(db *database.Database, containerMgr *container.Manager, logger *logrus.Logger) *Handler {
	return &Handler{
		db:           db,
		containerMgr: containerMgr,
		logger:       logger,
	}
}

func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api")

	switch {
	case path == "/functions" && r.Method == "GET":
		h.listFunctions(w, r)
	case path == "/functions" && r.Method == "POST":
		h.createFunction(w, r)
	case strings.HasPrefix(path, "/functions/") && r.Method == "GET":
		h.getFunction(w, r)
	case strings.HasPrefix(path, "/functions/") && r.Method == "PUT":
		h.updateFunction(w, r)
	case strings.HasPrefix(path, "/functions/") && r.Method == "DELETE":
		h.deleteFunction(w, r)
	case strings.HasSuffix(path, "/start") && r.Method == "POST":
		h.startFunction(w, r)
	case strings.HasSuffix(path, "/stop") && r.Method == "POST":
		h.stopFunction(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) listFunctions(w http.ResponseWriter, r *http.Request) {
	functions, err := h.db.ListFunctions()
	if err != nil {
		h.logger.Errorf("Failed to list functions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := make([]FunctionResponse, len(functions))
	for i, fn := range functions {
		response[i] = FunctionResponse{
			ID:        fn.ID,
			Name:      fn.Name,
			Language:  fn.Language,
			Code:      fn.Code,
			Port:      fn.Port,
			Status:    fn.Status,
			CreatedAt: fn.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: fn.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) createFunction(w http.ResponseWriter, r *http.Request) {
	var req CreateFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Language == "" || req.Code == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	port, err := h.db.GetAvailablePort()
	if err != nil {
		h.logger.Errorf("Failed to get available port: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	function := &database.Function{
		Name:     req.Name,
		Language: req.Language,
		Code:     req.Code,
		Port:     port,
		Status:   "stopped",
	}

	if err := h.db.CreateFunction(function); err != nil {
		h.logger.Errorf("Failed to create function: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(function)
}

func (h *Handler) getFunction(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/functions/")

	function, err := h.db.GetFunction(name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(function)
}

func (h *Handler) updateFunction(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/functions/")

	var req CreateFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	function, err := h.db.GetFunction(name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	function.Language = req.Language
	function.Code = req.Code

	if err := h.db.UpdateFunction(function); err != nil {
		h.logger.Errorf("Failed to update function: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(function)
}

func (h *Handler) deleteFunction(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/functions/")

	if err := h.db.DeleteFunction(name); err != nil {
		h.logger.Errorf("Failed to delete function: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) startFunction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/start")
	name := strings.TrimPrefix(path, "/api/functions/")

	if err := h.db.UpdateFunctionStatus(name, "starting"); err != nil {
		h.logger.Errorf("Failed to update function status: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) stopFunction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/stop")
	name := strings.TrimPrefix(path, "/api/functions/")

	if err := h.db.UpdateFunctionStatus(name, "stopped"); err != nil {
		h.logger.Errorf("Failed to update function status: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
