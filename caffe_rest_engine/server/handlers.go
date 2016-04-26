package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/boltdb/bolt"
)

var WorkQueue chan Job = make(chan Job)

const (
	DB_NAME = "models.db"
)

var MODELS_BUCKET = []byte("models")

type Job struct {
	Model  string
	Image  string //This can probably stay - we can just pass in base64 image strings into Caffe and have it decode.
	Output chan string
}

func init() {
	db, err := bolt.Open(DB_NAME, 0666, nil)
	if err != nil {
		panic("Failed to open database: " + err.Error())
	}

	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(MODELS_BUCKET)
		if err != nil {
			panic("Failed to create models bucket!")
		}
		return nil
	})
}

func writeResp(w http.ResponseWriter, resp interface{}, status int) {
	json, _ := json.Marshal(resp)
	w.WriteHeader(status)
	w.Write(json)
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

	select {
	case classified := <-unpackedJob.Output:
		fmt.Println("Request returning.")
		tmp := map[string]string{"output": string(classified)}
		marshalled, _ := json.Marshal(tmp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(marshalled)
		return
	}

}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var reg RegisterRequest
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	err := decoder.Decode(&reg)
	if err != nil {
		resp := RegisterResponse{
			Success: false,
			Error:   err.Error(),
		}

		writeResp(w, resp, http.StatusInternalServerError)
		return
	} else {
		panic("Not implemented. Implement DB code here.")
		for name, modelURL := range reg.Models {
		}

		modelKeys := make([]string, len(models))

		i := 0
		for k := range models {
			modelKeys[i] = k
			i++
		}

		resp := RegisterResponse{
			Success:   true,
			AllModels: modelKeys,
		}

		writeResp(w, resp, 200)
		return

	}

}
