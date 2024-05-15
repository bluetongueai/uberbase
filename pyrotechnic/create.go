package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

// TODO - research and document
var ipByte byte = 3

type CreateRequest struct {
	RootDrivePath string `json:"root_image_path"`
	KernelPath    string `json:"kernel_path"`
	CloudInitPath string `json:"cloud_init_path"`
}

type CreateResponse struct {
	IpAddress string `json:"ip_address"`
	ID        string `json:"id"`
}

// handle create vm request
// - create a vm
// - create a goroutine and wait on vm (basically a thread join?)
func CreateRequestHandler(w http.ResponseWriter, r *http.Request) {
	ipByte += 1
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("failed to read body, %s", err)
	}
	var req CreateRequest
	json.Unmarshal([]byte(body), &req)
	opts := GetOptions(ipByte, req)
	running, err := CreateVMM(opts, context.Background())
	if err != nil {
		log.Fatalf(err.Error())
	}

	id := Pseudo_uuid()
	resp := CreateResponse{
		IpAddress: opts.FcIP,
		ID:        id,
	}
	response, err := json.Marshal(&resp)
	if err != nil {
		log.Fatalf("failed to marshal json, %s", err)
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(response)

	Manager.Add(id, *running)

	go func() {
		defer running.cancelCtx()
		// there's an error here but we ignore it for now because we terminate
		// the VM on /delete and it returns an error when it's terminated
		running.machine.Wait(running.ctx)
	}()
}

// support cloud-init?
func makeIso(cloudInitPath string) (string, error) {
	image := "/tmp/cloud-init.iso"
	metaDataPath := "/tmp/my-meta-data.yml"
	err := os.WriteFile(metaDataPath, []byte("instance-id: i-litchi12345"), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create metadata file: %s", err)
	}
	if err := exec.Command("cloud-localds", image, cloudInitPath, metaDataPath).Run(); err != nil {
		return "", fmt.Errorf("cloud-localds failed: %s", err)
	}
	return image, nil
}
