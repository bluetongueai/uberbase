package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	f "github.com/tgittos/uberbase/functions/api/pkg/functions"
	h "github.com/tgittos/uberbase/functions/api/pkg/http"
)

type ApiConfig struct {
	Port  int      `json:"port"`
	Build string   `json:"build"`
	Pull  []string `json:"pull"`
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

	f.Init(f.FunctionsConfig{
		Build:  apiConfig.Build,
		Images: apiConfig.Pull,
	})

	s := h.NewServer(h.ServerConfig{
		Port: apiConfig.Port,
	})
	s.AddRoute("POST", "/api/v1/functions/*name", functionHandler)
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

func functionHandler(c *gin.Context) {
	name := strings.TrimPrefix(c.Param("name"), "/")

	params := c.PostForm("params")

	image_params := []string{}
	if params != "" {
		image_params = strings.Split(params, " ")
	}

	log.Printf("running function %s with params %v", name, image_params)

	output, err := f.Run(name, image_params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"output": output})
}
