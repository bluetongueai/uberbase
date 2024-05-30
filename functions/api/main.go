package main

import (
	"os"
  "net/http"

  "github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

func main() {
  r := gin.Default()
  config := Configure()

	// install middleware
	r.Use(Auth())
	r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{config.AllowedOrigins},
    AllowMethods:     []string{"*"},
    AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
    MaxAge: 12 * time.Hour,
  }))

  r.GET("/health", func(c *gin.Context) {
    c.JSON(http.StatusOK)
  })

	r.POST("/api/v1/functions/:name", func(c *gin.Context) {
    name := c.Param("name")
    // hammertime create -a 192.168.1.66:9090 -n my-microvm -ns my-namespace
    app := "hammertime"

    cmd := os.exec.Command(app, "create", "-a", "0.0.0.0:9090", "-n", name, "-ns" "uberbase_flintlock")

    stdout, err := cmd.Output()
	})

  r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
