package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type DeleteRequest struct {
	ID string `json:"id"`
}

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

	Manager.Shutdown(req.ID)
	Manager.Remove(req.ID)
}
