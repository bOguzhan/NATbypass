// internal/signaling/server.go
package signaling

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/bOguzhan/NATbypass/internal/utils"
	"github.com/gin-gonic/gin"
)

// Server represents the signaling HTTP server
type Server struct {
	router   *gin.Engine
	handlers *Handlers
	logger   *utils.Logger
	server   *http.Server
	mu       sync.Mutex
	clients  map[string]ClientInfo
}

// ClientInfo stores information about a connected client
type ClientInfo struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	IPAddress  string            `json:"ip_address"`
	LastSeen   time.Time         `json:"last_seen"`
	UserAgent  string            `json:"user_agent"`
	IsOnline   bool              `json:"is_online"`
	Properties map[string]string `json:"properties"`
}

// NewServer creates a new signaling server instance
func NewServer(logger *utils.Logger) *Server {
	// Create a new Gin router
	router := gin.New()
	router.Use(gin.Recovery())

	server := &Server{
		router:   router,
		logger:   logger,
		handlers: NewHandlers(logger),
		clients:  make(map[string]ClientInfo),
	}

	// Configure routes
	server.setupRoutes()

	return server
}

// setupRoutes configures all HTTP routes for the server
func (s *Server) setupRoutes() {
	// Add custom logger middleware
	s.router.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		start := time.Now()

		s.logger.WithFields(map[string]interface{}{
			"method": method,
			"path":   path,
			"client": c.ClientIP(),
			"ua":     c.Request.UserAgent(),
		}).Info("Request")

		c.Next()

		latency := time.Since(start)

		s.logger.WithFields(map[string]interface{}{
			"method":  method,
			"path":    path,
			"status":  c.Writer.Status(),
			"latency": latency,
			"size":    c.Writer.Size(),
		}).Debug("Response")
	})

	// Set up all routes
	s.handlers.SetupRoutes(s.router)

	// Add any server-specific routes
	s.router.GET("/stats", s.handleStats)
}

// handleStats returns server statistics
func (s *Server) handleStats(c *gin.Context) {
	s.mu.Lock()
	clientCount := len(s.clients)
	activeClients := 0

	for _, client := range s.clients {
		if client.IsOnline {
			activeClients++
		}
	}
	s.mu.Unlock()

	sysInfo, _ := utils.GetSystemInfo()

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"stats": gin.H{
			"total_clients":  clientCount,
			"active_clients": activeClients,
			"uptime_seconds": time.Since(sysInfo.StartTime).Seconds(),
			"version":        "0.1.0",
		},
	})
}

// Start begins the HTTP server on the specified address
func (s *Server) Start(address string) error {
	s.server = &http.Server{
		Addr:    address,
		Handler: s.router,
	}

	s.logger.Infof("Starting HTTP server on %s", address)
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

// RegisterClient registers a client with the server
func (s *Server) RegisterClient(id string, info ClientInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients[id] = info
	s.logger.Infof("Client registered: %s (%s)", id, info.Name)
}

// GetClient retrieves client information
func (s *Server) GetClient(id string) (ClientInfo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, exists := s.clients[id]
	return info, exists
}

// RemoveClient removes a client
func (s *Server) RemoveClient(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, id)
	s.logger.Infof("Client removed: %s", id)
}
