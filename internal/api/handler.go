package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/litefaas/litefaas/internal/container"
	"github.com/litefaas/litefaas/internal/database"
	"github.com/litefaas/litefaas/internal/proxy"
)

type Handler struct {
	db              *database.DB
	containerManager *container.Manager
	proxy           *proxy.Proxy
}

type CreateFunctionRequest struct {
	Name     string `json:"name"`
	Language string `json:"language"`
	Code     string `json:"code"`
}

type UpdateFunctionRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

type FunctionResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Language    string  `json:"language"`
	Code        string  `json:"code"`
	Port        *int    `json:"port"`
	Status      string  `json:"status"`
	ContainerID *string `json:"container_id"`
}

func NewHandler(db *database.DB, containerManager *container.Manager, proxy *proxy.Proxy) *Handler {
	return &Handler{
		db:              db,
		containerManager: containerManager,
		proxy:           proxy,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	switch parts[0] {
	case "functions":
		h.handleFunctions(w, r, parts[1:])
	case "health":
		h.handleHealth(w, r)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (h *Handler) handleFunctions(w http.ResponseWriter, r *http.Request, parts []string) {
	switch r.Method {
	case "GET":
		if len(parts) == 0 {
			h.listFunctions(w, r)
		} else {
			h.getFunction(w, r, parts[0])
		}
	case "POST":
		if len(parts) == 0 {
			h.createFunction(w, r)
		} else if len(parts) == 2 && parts[1] == "start" {
			h.startFunction(w, r, parts[0])
		} else if len(parts) == 2 && parts[1] == "stop" {
			h.stopFunction(w, r, parts[0])
		} else {
			http.Error(w, "Invalid path", http.StatusBadRequest)
		}
	case "PUT":
		if len(parts) == 1 {
			h.updateFunction(w, r, parts[0])
		} else {
			http.Error(w, "Invalid path", http.StatusBadRequest)
		}
	case "DELETE":
		if len(parts) == 1 {
			h.deleteFunction(w, r, parts[0])
		} else {
			http.Error(w, "Invalid path", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) listFunctions(w http.ResponseWriter, r *http.Request) {
	functions, err := h.db.ListFunctions()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list functions: %v", err), http.StatusInternalServerError)
		return
	}

	response := make([]FunctionResponse, len(functions))
	for i, f := range functions {
		response[i] = FunctionResponse{
			ID:          f.ID,
			Name:        f.Name,
			Language:    f.Language,
			Code:        f.Code,
			Port:        f.Port,
			Status:      f.Status,
			ContainerID: f.ContainerID,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) getFunction(w http.ResponseWriter, r *http.Request, name string) {
	function, err := h.db.GetFunctionByName(name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	response := FunctionResponse{
		ID:          function.ID,
		Name:        function.Name,
		Language:    function.Language,
		Code:        function.Code,
		Port:        function.Port,
		Status:      function.Status,
		ContainerID: function.ContainerID,
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
		http.Error(w, "Name, language, and code are required", http.StatusBadRequest)
		return
	}

	if req.Language != "python" && req.Language != "nodejs" {
		http.Error(w, "Unsupported language. Supported: python, nodejs", http.StatusBadRequest)
		return
	}

	function, err := h.db.CreateFunction(req.Name, req.Language, req.Code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create function: %v", err), http.StatusInternalServerError)
		return
	}

	response := FunctionResponse{
		ID:          function.ID,
		Name:        function.Name,
		Language:    function.Language,
		Code:        function.Code,
		Port:        function.Port,
		Status:      function.Status,
		ContainerID: function.ContainerID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) updateFunction(w http.ResponseWriter, r *http.Request, name string) {
	var req UpdateFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	function, err := h.db.GetFunctionByName(name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	updates := make(map[string]interface{})
	if req.Language != "" {
		if req.Language != "python" && req.Language != "nodejs" {
			http.Error(w, "Unsupported language. Supported: python, nodejs", http.StatusBadRequest)
			return
		}
		updates["language"] = req.Language
	}
	if req.Code != "" {
		updates["code"] = req.Code
	}

	if err := h.db.UpdateFunction(function.ID, updates); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update function: %v", err), http.StatusInternalServerError)
		return
	}

	updatedFunction, err := h.db.GetFunctionByID(function.ID)
	if err != nil {
		http.Error(w, "Failed to get updated function", http.StatusInternalServerError)
		return
	}

	response := FunctionResponse{
		ID:          updatedFunction.ID,
		Name:        updatedFunction.Name,
		Language:    updatedFunction.Language,
		Code:        updatedFunction.Code,
		Port:        updatedFunction.Port,
		Status:      updatedFunction.Status,
		ContainerID: updatedFunction.ContainerID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) deleteFunction(w http.ResponseWriter, r *http.Request, name string) {
	function, err := h.db.GetFunctionByName(name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	if function.Status == "running" {
		if err := h.proxy.StopFunction(r.Context(), name); err != nil {
			fmt.Printf("Failed to stop function before deletion: %v\n", err)
		}
	}

	if err := h.db.DeleteFunction(function.ID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete function: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) startFunction(w http.ResponseWriter, r *http.Request, name string) {
	if err := h.proxy.StartFunction(r.Context(), name); err != nil {
		// Log l'erreur pour le débogage
		fmt.Printf("Error starting function %s: %v\n", name, err)

		// Détecter si c'est le mode développement
		statusCode := http.StatusInternalServerError
		if err.Error() == "container management is disabled in development mode" {
			statusCode = http.StatusServiceUnavailable
		}

		// Retourner une réponse JSON avec plus de détails
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Failed to start function",
			"message": err.Error(),
			"function": name,
			"dev_mode": statusCode == http.StatusServiceUnavailable,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) stopFunction(w http.ResponseWriter, r *http.Request, name string) {
	if err := h.proxy.StopFunction(r.Context(), name); err != nil {
		http.Error(w, fmt.Sprintf("Failed to stop function: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}
