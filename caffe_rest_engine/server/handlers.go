package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/boltdb/bolt"
)

var WorkQueue chan Job = make(chan Job)
var db *bolt.DB

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
	var err error
	db, err = bolt.Open(DB_NAME, 0666, nil)
	if err != nil {
		panic("Failed to open database: " + err.Error())
	}

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
	w.Header().Set("Content-Type", "application/json")
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
		writeResp(w, tmp, 200)
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

		modelArray := make([]Model, len(reg.Models))

		i := 0
		for name, modelURL := range reg.Models {
			model := NewModelFromURL(name, modelURL)
			modelArray[i] = model
			i++
		}

		for model := range modelArray {
			var buf bytes.Buffer
			db.Update(func(tx *bolt.Tx) error {

				b := tx.Bucket(MODELS_BUCKET)
				err = b.Put([]byte(name), []byte(modelURL))
				if err != nil {
					return err
				}
				return nil
			})
		}
		modelKeys := make([]string, 0)

		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(MODELS_BUCKET)
			c := b.Cursor()

			for k, v := c.First(); k != nil; k, v = c.Next() {
				modelKeys = append(modelKeys, string(k))
			}

			return nil
		})

		resp := RegisterResponse{
			Success:   true,
			AllModels: modelKeys,
		}

		writeResp(w, resp, 200)
		return

	}

}
