package main

// #include <stdlib.h>
// #include <classification.h>
import "C"
import "unsafe"
import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"net/http"
)

var WorkQueue chan Job = make(chan Job)
var db *bolt.DB

const (
	DB_NAME = "models.db"
)

var MODELS_BUCKET = []byte("models")

type Job struct {
	Model  string
	Image  C.c_mat //This can probably stay - we can just pass in base64 image strings into Caffe and have it decode.
	Output chan string
}

type TempJob struct {
	Model string
	Image string
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
	if r.Method != "POST" {
		http.Error(w, "Invalid method - only POST requests are valid for this endpoint.", 405)
	}
	var unpackedRes TempJob
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&unpackedRes)
	defer r.Body.Close()
	data, err := base64.StdEncoding.DecodeString(unpackedRes.Image)
	if err != nil {
		fmt.Println("Base64 error " + err.Error())
		return
	}
	job := Job{
		Model: unpackedRes.Model,
		Image: C.make_mat((*C.char)(unsafe.Pointer(&data[0])), C.size_t(len(data))),
	}
	if err != nil {
		http.Error(w, "Invalid JSON "+err.Error(), http.StatusBadRequest)
		return
	}

	job.Output = make(chan string)
	fmt.Println("\nAdded to work queue.")
	WorkQueue <- job

	select {
	case classified := <-job.Output:
		fmt.Println("Request returning.")
		writeResp(w, classified, 200)
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

}
