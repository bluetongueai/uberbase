package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"fmt"
	"syscall"
	"strings"
	"io"

	"github.com/gin-gonic/gin"
	f "github.com/tgittos/uberbase/functions/api/pkg/functions"
	h "github.com/tgittos/uberbase/functions/api/pkg/http"
)

type ApiConfig struct {
	Port  int      `json:"port"`
	Build string   `json:"build"`
	Pull  []string `json:"pull"`
}

type FunctionRequest struct {
	Args		*[]string 	`json:args`
	Detatch *bool			`json:detatch`
}

type StopRequest struct {
	ContainerId string `json:containerId`
}

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatalf("config file path not provided")
	}
	configPath := args[1]
	apiConfig, err := readConfigFile(configPath)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	// capture sigs
	log.Printf("hooking into OS signals to gracefully shutdown")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Printf("received signal, shutting down")
		f.Shutdown()
		os.Exit(0)
	}()

	f.Init(f.FunctionsConfig{
		Build:  apiConfig.Build,
		Pull: apiConfig.Pull,
	})

	s := h.NewServer(h.ServerConfig{
		Port: apiConfig.Port,
	})
	s.AddRoute("POST", "/api/v1/functions/stop", stopHandler)
	s.AddRoute("POST", "/api/v1/functions/run/*name", functionHandler)
	s.Start()
}

func readConfigFile(configPath string) (ApiConfig, error) {
	// read file at ./config.json
	configFileBytes, _ := os.ReadFile(configPath)
	log.Printf("config file: %s", string(configFileBytes))
	var data ApiConfig
	err := json.Unmarshal(configFileBytes, &data)
	if err != nil {
		log.Printf("reading config file failed: %v", err)
		return data, err
	}
	return data, nil
}

func stopHandler(c *gin.Context) {
	var request StopRequest
  err := c.BindJSON(&request);
	if err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid JSON request: %v", err),
		})
		return
	}

	stdout, stderr, err := f.Stop(request.ContainerId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failure",
			"error": err.Error(),
			"stdout": stdout,
			"stderr": stderr,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"stdout": stdout,
		"stderr": stderr,
	})
}

func functionHandler(c *gin.Context) {
	name := strings.TrimPrefix(c.Param("name"), "/")

	var request FunctionRequest
  err := c.BindJSON(&request);
	if err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid JSON request: %v", err),
		})
		return
	}

	args := []string{}
	if request.Args != nil {
		args = *request.Args
	}
	log.Printf("running image %s with args %v", name, args)
	detatch := false
	if request.Detatch != nil {
		detatch = *request.Detatch
	}
	stdout, stderr, err := f.Run(name, detatch, args...)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failure",
			"error": err.Error(),
			"stdout": stdout,
			"stderr": stderr,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"stdout": stdout,
		"stderr": stderr,
	})
}
