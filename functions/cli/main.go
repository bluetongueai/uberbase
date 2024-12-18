package main

import (
	"log"
	"os"

	"github.com/tgittos/uberbase/functions/cli/pkg/cli"
)

func main() {
	c := cli.NewCLI()
	if err := c.Execute(); err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
