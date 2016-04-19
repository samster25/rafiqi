package main

import (
	"fmt"
	"net/http"
	"encoding/json"
)

var WorkQueue chan Job = make(chan Job)

type Job struct {
	Model string
	Image string //This can probably stay - we can just pass in base64 image strings into Caffe and have it decode. 
	Output chan string
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
	unpackedJob.Output = make(chan string)
	fmt.Println("\nAdded to work queue.")
	WorkQueue <- unpackedJob
	for {
		select {
		case classified := <- unpackedJob.Output:
			fmt.Println("Request returning.")
			tmp := map[string]string{"output": string(classified)}
			marshalled, _ := json.Marshal(tmp)
			w.Header().Set("Content-Type", "application/json")
  			w.Write(marshalled)
  			return
		}
	}
}
