package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var Manager = FirecrackerMVMManager{
	RunningVMs: make(map[string]RunningFirecracker),
}

// start an API server at port 8080
// and offer two urls: create, delete
func main() {
	f, _ := CreateLogFile()
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	host := os.Getenv("PYROTECHNIC_HOST")
	port := os.Getenv("PYROTECHNIC_PORT")
	exposed := fmt.Sprintf("%s:%s", host, port)
	http.HandleFunc("/create", CreateRequestHandler)
	http.HandleFunc("/delete", DeleteRequestHandler)

	defer Manager.ShutdownAll()

	log.Printf("pyrotechnic server starting at %s", exposed)
	log.Fatal(http.ListenAndServe(exposed, nil))
}
