package main

import (
	"fmt"
	"net/http"
	"encoding/json"
)

var WorkQueue chan Job = make(chan Job)

type Job struct {
	Model string
	Image string
	w http.ResponseWriter
}

func JobHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid method - only POST requests are valid for this endpoint.", 405)
	}
	var unpackedJob Job
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&unpackedJob)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Invalid JSON "+err.Error(), http.StatusBadRequest)
    	return
	}
	//unpackedJob.w = w
	fmt.Println("Added to work queue.\n")
	WorkQueue <- unpackedJob
	w.WriteHeader(http.StatusCreated)
	return

}
