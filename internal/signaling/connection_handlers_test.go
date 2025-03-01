// internal/signaling/connection_handlers_test.go
package signaling

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bOguzhan/NATbypass/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupConnectionTestRouter() (*gin.Engine, *Handlers) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := utils.NewLogger("test", "info")
	handlers := NewHandlers(logger)

	// Initialize the connection registry
	handlers.connections = NewConnectionRegistry(logger)

	// Set up routes
	handlers.SetupRoutes(router)

	return router, handlers
}

func TestRequestConnection(t *testing.T) {
	router, _ := setupConnectionTestRouter() // Using _ to ignore the handlers variable

	// 1. Test valid connection request
	w := httptest.NewRecorder()
	reqBody := map[string]interface{}{
		"source_id":   "12345678901234567890123456789012",
		"target_id":   "21098765432109876543210987654321",
		"source_ip":   "192.168.1.5",
		"source_port": 12345,
	}
	reqJSON, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/connect", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "connection_registered", response["status"])
	assert.NotEmpty(t, response["connection_id"])

	// Capture the connection ID for later tests
	connID := response["connection_id"].(string)

	// 2. Test getting connections for client
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/connections/12345678901234567890123456789012", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var connResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &connResponse)
	assert.NoError(t, err)
	assert.Equal(t, "success", connResponse["status"])

	connections := connResponse["connections"].([]interface{})
	assert.Len(t, connections, 1)
	assert.Equal(t, connID, connections[0].(map[string]interface{})["connection_id"])
	assert.Equal(t, "initiated", connections[0].(map[string]interface{})["status"])

	// 3. Test updating connection status
	w = httptest.NewRecorder()
	updateReq := map[string]interface{}{
		"connection_id": connID,
		"status":        "negotiating",
	}
	updateJSON, _ := json.Marshal(updateReq)
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/connection/update", bytes.NewBuffer(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the status was updated
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/connections/12345678901234567890123456789012", nil)
	router.ServeHTTP(w, req)

	err = json.Unmarshal(w.Body.Bytes(), &connResponse)
	assert.NoError(t, err)
	connections = connResponse["connections"].([]interface{})
	assert.Equal(t, "negotiating", connections[0].(map[string]interface{})["status"])

	// 4. Test connection error
	w = httptest.NewRecorder()
	errorReq := map[string]interface{}{
		"connection_id": connID,
		"status":        "failed",
		"error_message": "ICE negotiation timeout",
	}
	errorJSON, _ := json.Marshal(errorReq)
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/connection/update", bytes.NewBuffer(errorJSON))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Test invalid request
	w = httptest.NewRecorder()
	invalidReq := map[string]interface{}{
		"source_id": "too-short",
		"target_id": "also-too-short",
	}
	invalidJSON, _ := json.Marshal(invalidReq)
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/connect", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetNonExistentConnection(t *testing.T) {
	router, _ := setupConnectionTestRouter()

	// Use a valid format client ID (32 characters)
	// The handler is rejecting the ID because it's not the right length
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/connections/12345678901234567890123456789012", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check that we got a success status
	assert.Equal(t, "success", response["status"])

	// Safely check if connections exist and are empty
	connections, exists := response["connections"]
	assert.True(t, exists, "Response should contain 'connections' field")

	// Now safely convert to a slice
	connSlice, ok := connections.([]interface{})
	assert.True(t, ok, "Connections should be a slice")
	assert.Empty(t, connSlice, "Connections slice should be empty")
}

func TestUpdateNonExistentConnection(t *testing.T) {
	router, _ := setupConnectionTestRouter() // Using _ to ignore the handlers variable

	w := httptest.NewRecorder()
	updateReq := map[string]interface{}{
		"connection_id": "nonexistent-conn-id",
		"status":        "established",
	}
	updateJSON, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/connection/update", bytes.NewBuffer(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "connection_not_found", response["error"])
}
