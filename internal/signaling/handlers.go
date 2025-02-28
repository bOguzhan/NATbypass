// internal/signaling/handlers.go

package signaling

import (
	"net/http"
	"time"

	"github.com/bOguzhan/NATbypass/internal/config"
	"github.com/bOguzhan/NATbypass/internal/utils"
	"github.com/bOguzhan/NATbypass/pkg/networking"
	"github.com/gin-gonic/gin"
)

// Handlers encapsulates the HTTP handlers for the signaling server
type Handlers struct {
	logger *utils.Logger
	server *Server // Reference to the server for client management
	config *config.Config
}

// NewHandlers creates a new instance of signaling handlers
func NewHandlers(logger *utils.Logger) *Handlers {
	return &Handlers{
		logger: logger,
	}
}

// SetServer sets the server reference
func (h *Handlers) SetServer(server *Server) {
	h.server = server
}

// SetConfig sets the configuration reference
func (h *Handlers) SetConfig(cfg *config.Config) {
	h.config = cfg
}

// RegisterClient handles client registration requests
func (h *Handlers) RegisterClient(c *gin.Context) {
	type RegisterRequest struct {
		ClientID   string            `json:"client_id"`
		Name       string            `json:"name"`
		Properties map[string]string `json:"properties"`
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

	// Store client information if server is available
	if h.server != nil {
		h.server.RegisterClient(clientID, ClientInfo{
			ID:         clientID,
			Name:       req.Name,
			IPAddress:  c.ClientIP(),
			LastSeen:   time.Now(),
			UserAgent:  c.Request.UserAgent(),
			IsOnline:   true,
			Properties: req.Properties,
		})
	}

	h.logger.WithFields(map[string]interface{}{
		"client_id": clientID,
		"name":      req.Name,
		"ip":        c.ClientIP(),
	}).Info("Client registered")

	c.JSON(http.StatusOK, gin.H{
		"status":    "registered",
		"client_id": clientID,
		"timestamp": time.Now(),
	})
}

// GetPublicAddress returns the client's public IP and port using STUN
func (h *Handlers) GetPublicAddress(c *gin.Context) {
	// Get client's IP from headers or connection
	clientIP := c.ClientIP()

	// Use STUN to determine actual public address
	stunServer := "stun.l.google.com:19302" // Default STUN server

	// Use config if available
	if h.config != nil && h.config.Stun.Server != "" {
		stunServer = h.config.Stun.Server
	}

	h.logger.Debug("Attempting STUN discovery for client address")

	// Create STUN config
	stunConfig := networking.STUNConfig{
		Server:         stunServer,
		TimeoutSeconds: 5,
		RetryCount:     3,
	}

	// Use config values if available
	if h.config != nil {
		stunConfig.TimeoutSeconds = h.config.Stun.TimeoutSeconds
		stunConfig.RetryCount = h.config.Stun.RetryCount
	}

	addr, err := networking.DiscoverPublicAddressWithConfig(stunConfig)

	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"error":     err.Error(),
			"client_ip": clientIP,
		}).Warn("Failed to discover public address via STUN")

		// Fall back to client IP from HTTP headers
		c.JSON(http.StatusOK, gin.H{
			"status":    "partial",
			"ip":        clientIP,
			"message":   "STUN discovery failed, using HTTP-derived IP",
			"timestamp": time.Now(),
		})
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"client_ip": clientIP,
		"stun_ip":   addr.IP.String(),
		"stun_port": addr.Port,
	}).Debug("Public address discovered via STUN")

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"ip":        addr.IP.String(),
		"port":      addr.Port,
		"timestamp": time.Now(),
	})
}

// Heartbeat handles client heartbeat to keep connection alive
func (h *Handlers) Heartbeat(c *gin.Context) {
	type HeartbeatRequest struct {
		ClientID string `json:"client_id"`
	}

	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ClientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "invalid_request",
		})
		return
	}

	// Update client last seen time if server is available
	if h.server != nil {
		if info, exists := h.server.GetClient(req.ClientID); exists {
			info.LastSeen = time.Now()
			info.IsOnline = true
			h.server.RegisterClient(req.ClientID, info) // Update info

			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"timestamp": time.Now(),
			})
			return
		}

		// Client ID not found
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"error":   "client_not_found",
			"message": "The specified client ID is not registered",
		})
		return
	}

	// Server reference not available
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"message":   "Heartbeat received, but client tracking is not available",
		"timestamp": time.Now(),
	})
}

// SetupRoutes configures all the routes for the signaling server
func (h *Handlers) SetupRoutes(router *gin.Engine) {
	// API v1 group
	v1 := router.Group("/api/v1")
	{
		v1.POST("/register", h.RegisterClient)
		v1.GET("/address", h.GetPublicAddress)
		v1.POST("/heartbeat", h.Heartbeat)

		// Client connections endpoints - will be implemented later
		v1.POST("/connect", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented"})
		})

		v1.POST("/signal", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{"status": "not_implemented"})
		})
	}

	// Version info
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version": "0.1.0",
			"name":    "mediatory-server",
		})
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "mediatory-server",
			"time":    time.Now(),
		})
	})
}
