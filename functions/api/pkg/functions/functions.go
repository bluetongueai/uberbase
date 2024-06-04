package functions

import (
	"context"
	"errors"
	"log"
	"os"
)

type FunctionsConfig struct {
	MinPoolSize int
	MaxPoolSize int
	Images      []string
}

var containerContext context.Context
var pool containerPool
var initialized bool

func Init(config FunctionsConfig) error {
	log.SetOutput(os.Stdout)

	poolConfig := containerPoolConfig{
		InitialSize: config.MinPoolSize,
		MaxSize:     config.MaxPoolSize,
		Images:      config.Images,
	}

	log.Printf("initializing container pool with min size %d and max size %d", config.MinPoolSize, config.MaxPoolSize)
	p, err := newContainerPool(poolConfig)
	if err != nil {
		return err
	}
	pool = p
	initialized = true
	log.Println("container pool initialized")
	return nil
}

func Run(imageName string, params ...string) (string, error) {
	if !initialized {
		return "", errors.New("functions not initialized")
	}
	return pool.Run(imageName, params...)
}
