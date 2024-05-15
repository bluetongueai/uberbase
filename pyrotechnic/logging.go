package main

import (
	"log"
	"os"
)

func CreateLogFile() (*os.File, error) {
	f, err := os.OpenFile("/var/log/pyrotechnic.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	return f, nil
}
