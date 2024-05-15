package main

import (
	"io"
	"json"
	"log"
	"os"
)

// handle delete request
// - shutdown the running vm
// - delete it
func DeleteRequestHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("failed to read body, %s", err)
	}
	var req DeleteRequest
	json.Unmarshal([]byte(body), &req)
	if err != nil {
		log.Fatalf(err.Error())
	}

	running := runningVMs[req.ID]
	shutDown(running)
	delete(runningVMs, req.ID)
}

// shutdown a running vm
func ShutDown(running RunningFirecracker) {
	running.machine.StopVMM()
	os.Remove(running.image)
}

