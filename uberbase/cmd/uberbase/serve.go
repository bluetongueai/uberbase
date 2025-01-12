package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	f "github.com/bluetongueai/uberbase/uberbase/pkg/functions"
	h "github.com/bluetongueai/uberbase/uberbase/pkg/http"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

type ApiConfig struct {
	Port  int      `json:"port"`
	Build string   `json:"build"`
	Pull  []string `json:"pull"`
}

type FunctionRequest struct {
	Args    *[]string          `json:args`
	Detatch *bool              `json:detatch`
	Env     *map[string]string `json:env`
}

type StopRequest struct {
	ContainerId string `json:containerId`
}

func getServeCmd() *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve [config-file]",
		Short: "Start the Uberbase API server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := args[0]
			apiConfig, err := readConfigFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to read config file: %v", err)
			}

			f.Init(f.FunctionsConfig{
				Build: apiConfig.Build,
				Pull:  apiConfig.Pull,
			})

			s := h.NewServer(h.ServerConfig{
				Port: apiConfig.Port,
			})
			s.AddRoute("POST", "/api/v1/functions/stop", stopHandler)
			s.AddRoute("POST", "/api/v1/functions/run/*name", functionHandler)
			s.Start()
			return nil
		},
	}

	return serveCmd
}

func readConfigFile(configPath string) (ApiConfig, error) {
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
	err := c.BindJSON(&request)
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
			"error":  err.Error(),
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
	err := c.BindJSON(&request)
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
	env := map[string]string{}
	if request.Env != nil {
		env = *request.Env
	}
	stdout, stderr, err := f.Run(name, detatch, env, args...)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failure",
			"error":  err.Error(),
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
