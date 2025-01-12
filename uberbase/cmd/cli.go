package main

import (
	"log"

	"github.com/bluetongueai/uberbase/uberbase/pkg/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}
