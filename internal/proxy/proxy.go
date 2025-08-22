package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/litefaas/litefaas/internal/container"
	"github.com/litefaas/litefaas/internal/database"
)

type Proxy struct {
	db              *database.DB
	containerManager *container.Manager
	portStart       int
	portEnd         int
	client          *http.Client
}

func New(db *database.DB, containerManager *container.Manager, portStart, portEnd int) *Proxy {
	return &Proxy{
		db:              db,
		containerManager: containerManager,
		portStart:       portStart,
		portEnd:         portEnd,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	functionName := r.URL.Path[1:] // Remove leading slash

	if functionName == "" {
		http.Error(w, "Function name required", http.StatusBadRequest)
		return
	}

	function, err := p.db.GetFunctionByName(functionName)
	if err != nil {
		fmt.Printf("Function not found: %s\n", functionName)
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	if function.Status != "running" {
		fmt.Printf("Function %s is not running (status: %s)\n", functionName, function.Status)
		http.Error(w, "Function is not running", http.StatusServiceUnavailable)
		return
	}

	if function.Port == nil || *function.Port == 0 {
		fmt.Printf("Function %s has no assigned port\n", functionName)
		http.Error(w, "Function has no assigned port", http.StatusInternalServerError)
		return
	}

	targetURL := fmt.Sprintf("http://127.0.0.1:%d%s", *function.Port, r.URL.Path)

	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	req.Header = r.Header
	req.URL.RawQuery = r.URL.RawQuery

	resp, err := p.client.Do(req)
	if err != nil {
		fmt.Printf("Failed to proxy request to function %s: %v\n", functionName, err)
		http.Error(w, "Function execution failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		fmt.Printf("Failed to copy response body: %v\n", err)
	}
}

func (p *Proxy) StartFunction(ctx context.Context, functionName string) error {
	if p.containerManager == nil {
		return fmt.Errorf("container management is disabled in development mode")
	}

	function, err := p.db.GetFunctionByName(functionName)
	if err != nil {
		return fmt.Errorf("function not found: %w", err)
	}

	if function.Status == "running" {
		return fmt.Errorf("function is already running")
	}

	port, err := p.db.GetAvailablePort()
	if err != nil {
		return fmt.Errorf("no available ports: %w", err)
	}

	containerID, err := p.containerManager.CreateFunctionContainer(ctx, function.Name, function.Language, function.Code, port)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	updates := map[string]interface{}{
		"port":         port,
		"status":       "running",
		"container_id": containerID,
	}

	if err := p.db.UpdateFunction(function.ID, updates); err != nil {
		p.containerManager.StopContainer(ctx, containerID)
		return fmt.Errorf("failed to update function status: %w", err)
	}

	fmt.Printf("Started function %s on port %d\n", functionName, port)
	return nil
}

func (p *Proxy) StopFunction(ctx context.Context, functionName string) error {
	function, err := p.db.GetFunctionByName(functionName)
	if err != nil {
		return fmt.Errorf("function not found: %w", err)
	}

	if function.Status != "running" {
		return fmt.Errorf("function is not running")
	}

	if function.ContainerID != nil && *function.ContainerID != "" && p.containerManager != nil {
		if err := p.containerManager.StopContainer(ctx, *function.ContainerID); err != nil {
			fmt.Printf("Failed to stop container %s: %v\n", *function.ContainerID, err)
		}
	}

	updates := map[string]interface{}{
		"status":       "stopped",
		"container_id": "",
	}

	if err := p.db.UpdateFunction(function.ID, updates); err != nil {
		return fmt.Errorf("failed to update function status: %w", err)
	}

	fmt.Printf("Stopped function %s\n", functionName)
	return nil
}

func (p *Proxy) GetFunctionStatus(functionName string) (string, error) {
	function, err := p.db.GetFunctionByName(functionName)
	if err != nil {
		return "", fmt.Errorf("function not found: %w", err)
	}

	if function.Status == "running" && function.ContainerID != nil && *function.ContainerID != "" && p.containerManager != nil {
		ctx := context.Background()
		containerStatus, err := p.containerManager.GetContainerStatus(ctx, *function.ContainerID)
		if err != nil {
			fmt.Printf("Failed to get container status: %v\n", err)
		} else if containerStatus != "running" {
			updates := map[string]interface{}{
				"status":       "stopped",
				"container_id": "",
			}
			p.db.UpdateFunction(function.ID, updates)
			return "stopped", nil
		}
	}

	return function.Status, nil
}
