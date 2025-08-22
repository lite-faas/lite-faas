package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/litefaas/litefaas/internal/api"
	"github.com/litefaas/litefaas/internal/database"
	"github.com/litefaas/litefaas/internal/proxy"
)

func TestHealthEndpoint(t *testing.T) {
	// Créer une base de données temporaire
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Créer un proxy mock
	proxy := &proxy.Proxy{}

	// Créer le handler API
	handler := api.NewHandler(db, nil, proxy)

	// Créer une requête de test
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Créer un ResponseRecorder pour enregistrer la réponse
	rr := httptest.NewRecorder()

	// Appeler le handler
	handler.ServeHTTP(rr, req)

	// Vérifier le code de statut
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Vérifier le contenu de la réponse
	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response["status"])
	}
}
