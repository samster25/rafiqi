package main

// #include <stdlib.h>
// #include "classification.h"
import "C"
import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"net/http"
	"time"
)

var db *bolt.DB

const (
	DB_NAME = "models.db"
)

var MODELS_BUCKET = []byte("models")

type Job struct {
	Model  string
	Image  []byte
	Output chan string
}

type TempJob struct {
	Model string
	Image []byte
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

func writeError(w http.ResponseWriter, err error) {
	resp := RegisterResponse{
		Success: false,
		Error:   err.Error(),
	}

	writeResp(w, resp, http.StatusInternalServerError)
}

func JobHandler(w http.ResponseWriter, r *http.Request) {
	Debugf("Request beginning")
	start := time.Now()
	if r.Method != "POST" {
		http.Error(w, "Invalid method - only POST requests are valid for this endpoint.", 405)
	}
	image, err := ioutil.ReadAll(r.Body)
	if err != nil {
		handleError("Error reading image", err)
		return
	}
	defer r.Body.Close()

	loadedModels.RLock()
	job := Job{
		Model: r.FormValue("model_name"),
		Image: image,
	}
	loadedModels.RUnlock()
	job.Output = make(chan string)
	WorkQueue.AddJob(job)

	select {
	case classified := <-job.Output:
		writeResp(w, classified, 200)
		LogTimef("Request returning success", start)
		return
	}

}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var reg RegisterRequest
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	err := decoder.Decode(&reg)
	if err != nil {
		writeError(w, err)
	} else {

		modelArray := make([]Model, len(reg.Models))

		i := 0
		for name, modelReq := range reg.Models {
			model := NewModelFromURL(name, modelReq)
			modelArray[i] = model
			i++
		}

		for _, model := range modelArray {
			var encModel bytes.Buffer
			enc := gob.NewEncoder(&encModel)

			err := db.Update(func(tx *bolt.Tx) error {
				err := enc.Encode(model)
				if err != nil {
					return err
				}

				b := tx.Bucket(MODELS_BUCKET)
				err = b.Put([]byte(model.Name), encModel.Bytes())
				if err != nil {
					return err
				}

				encModel.Truncate(0)
				return nil
			})

			if err != nil {
				writeError(w, err)
				return
			}
			InitializeModel(&model)
		}

		resp := RegisterResponse{
			Success: true,
		}

		writeResp(w, resp, 200)
		return

	}

}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Invalid Method: Only GET requests allowed to this endpoint.", 405)
	}
	modelKeys := make([]string, 0)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MODELS_BUCKET)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			buf := bytes.NewBuffer(v)
			dec := gob.NewDecoder(buf)
			var decModel Model
			dec.Decode(&decModel)
			s := fmt.Sprintf("%v|%v|%v|%v|%v", string(k), decModel.WeightsPath,
				decModel.ModelPath, decModel.LabelsPath, decModel.MeanPath)
			modelKeys = append(modelKeys, s)
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
