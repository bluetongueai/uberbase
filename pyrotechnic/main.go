package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	models "github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

var runningVMs map[string]RunningFirecracker = 
// TODO - research and document
var ipByte byte = 3

// start an API server at port 8080
// and offer two urls: create, delete
func main() {
	http.HandleFunc("/create", CreateRequestHandler)
	http.HandleFunc("/delete", DeleteRequestHandler)
	defer cleanup()

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// shutdown all vms
func cleanup() {
	for _, running := range runningVMs {
		ShutDown(running)
	}
}
