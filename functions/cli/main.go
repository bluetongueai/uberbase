package main

import (
	"log"

	"github.com/tgittos/uberbase/functions/cli/pkg/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}
