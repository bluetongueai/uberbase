package http

import (
	"github.com/gin-gonic/gin"
)

func NewServer() Server {
	server := Server{
		gin: gin.Default(),
	}

	server.RegisterGlobalMiddleware(CorsMiddleware())
	server.AddRoute("GET", "/health", HealthHandler)

	return server
}

type Server struct {
	gin *gin.Engine
}

func (s Server) AddRoute(method string, path string, handler func(c *gin.Context)) {
	s.gin.Handle(method, path, handler)
}

func (s Server) Start() {
	s.gin.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func (s Server) RegisterGlobalMiddleware(middleware ...gin.HandlerFunc) {
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
