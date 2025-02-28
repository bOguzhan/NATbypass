// internal/signaling/handlers.go

package signaling

import (
	"net/http"

	"github.com/bOguzhan/NATbypass/internal/utils"
	"github.com/gin-gonic/gin"
)

// Handlers encapsulates the HTTP handlers for the signaling server
type Handlers struct {
	logger *utils.Logger
}

// NewHandlers creates a new instance of signaling handlers
func NewHandlers(logger *utils.Logger) *Handlers {
	return &Handlers{
		logger: logger,
	}
}

// RegisterClient handles client registration requests
func (h *Handlers) RegisterClient(c *gin.Context) {
	type RegisterRequest struct {
		ClientID string `json:"client_id"`
		Name     string `json:"name"`
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Warn("Invalid registration request")

		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "invalid_request",
			"message": "Invalid request format",
		})
		return
	}

	// Validate client ID if provided, or generate a new one
	clientID := req.ClientID
	if clientID == "" || !utils.ValidateID(clientID, 32) {
		var err error
		clientID, err = utils.GeneratePeerID()
		if err != nil {
			h.logger.WithFields(map[string]interface{}{
				"error": err.Error(),
			}).Error("Failed to generate peer ID")

			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"error":   "server_error",
				"message": "Failed to generate client ID",
			})
			return
		}
	}

	// TODO: Store client information in registry

	h.logger.WithFields(map[string]interface{}{
		"client_id": clientID,
		"name":      req.Name,
	}).Info("Client registered")

	c.JSON(http.StatusOK, gin.H{
		"status":    "registered",
		"client_id": clientID,
	})
}

// GetPublicAddress returns the client's public IP and port
func (h *Handlers) GetPublicAddress(c *gin.Context) {
	// Get client's IP from headers or connection
	clientIP := c.ClientIP()

	// TODO: Use STUN to determine actual public address
	// For now, just return what we have

	h.logger.WithFields(map[string]interface{}{
		"client_ip": clientIP,
	}).Debug("Public address requested")

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"ip":     clientIP,
		// Port will be added when STUN is implemented
	})
}

// SetupRoutes configures all the routes for the signaling server
func (h *Handlers) SetupRoutes(router *gin.Engine) {
	// API v1 group
	v1 := router.Group("/api/v1")
	{
		v1.POST("/register", h.RegisterClient)
		v1.GET("/address", h.GetPublicAddress)

		// Client connections endpoints - will be implemented later
		v1.POST("/connect", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented"})
		})

		v1.POST("/signal", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented"})
		})
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "mediatory-server",
			"version": "0.1.0",
		})
	})
}
