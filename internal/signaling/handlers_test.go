// internal/signaling/handlers_test.go
package signaling

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bOguzhan/NATbypass/internal/utils"
	"github.com/gin-gonic/gin"
)

func setupTestRouter() (*gin.Engine, *Handlers) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := utils.NewLogger("test", "info")
	handlers := NewHandlers(logger)
	handlers.SetupRoutes(router)
	return router, handlers
}

func TestHealthEndpoint(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if status, exists := response["status"]; !exists || status != "ok" {
		t.Errorf("Expected status 'ok', got %v", status)
	}
}

func TestRegisterClient(t *testing.T) {
	router, _ := setupTestRouter()

	// Test valid registration
	w := httptest.NewRecorder()
	reqBody := map[string]interface{}{
		"name": "test-client",
		"properties": map[string]string{
			"device": "laptop",
			"os":     "linux",
		},
	}
	reqJSON, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if status, exists := response["status"]; !exists || status != "registered" {
		t.Errorf("Expected status 'registered', got %v", status)
	}

	if _, exists := response["client_id"]; !exists {
		t.Error("Response should contain client_id")
	}
}

func TestGetPublicAddress(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/address", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if status, exists := response["status"]; !exists || status != "success" {
		t.Errorf("Expected status 'success', got %v", status)
	}

	if _, exists := response["ip"]; !exists {
		t.Error("Response should contain ip")
	}
}
