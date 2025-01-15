package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bluetongueai/uberbase/uberbase/pkg/logging"
	"github.com/gin-gonic/gin"
)

type ServerConfig struct {
	Port int `json:"port"`
}

var config ServerConfig

func NewServer(c ServerConfig) Server {
	config = c
	server := Server{
		gin: gin.Default(),
	}

	server.RegisterGlobalMiddleware(CorsMiddleware())
	server.AddRoute("GET", "/health", HealthHandler)

	return server
}

type Server struct {
	gin    *gin.Engine
	server *http.Server
}

func (s *Server) AddRoute(method string, path string, handler func(c *gin.Context)) {
	s.gin.Handle(method, path, handler)
}

func (s *Server) Start() {
	s.server = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", config.Port),
		Handler: s.gin,
	}

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logging.Logger.Errorf("Failed to start server: %v", err)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

func (s *Server) RegisterGlobalMiddleware(middleware ...gin.HandlerFunc) {
	s.gin.Use(middleware...)
}

// Add your routes here
/*
	r.GET("/protected", func(c *gin.Context) {
		user := c.MustGet("user").(string)
		fmt.Println("User:", user)
		c.String(http.StatusOK, "Hello, "+user+"!")
	})
*/
