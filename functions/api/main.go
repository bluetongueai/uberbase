package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	f "github.com/tgittos/uberbase/functions/api/pkg/functions"
	h "github.com/tgittos/uberbase/functions/api/pkg/http"
)

func main() {
	f.Init(f.FunctionsConfig{
		MinPoolSize: 60,
		MaxPoolSize: 300,
		Images: 		 []string{"docker.io/bluetongueai/functions-hello-world:latest"},
	})

	s := h.NewServer()
	s.AddRoute("POST", "/api/v1/functions/:name", functionHandler)
	s.Start()
}

func functionHandler(c *gin.Context) {
	name := c.Param("name")

	var http_params map[string]string
	if err := c.BindJSON(&http_params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	image_params := strings.Split(http_params["params"], " ")

	output, err := f.Run(name, image_params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"output": output})
}
