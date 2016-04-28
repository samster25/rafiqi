package main

// #cgo pkg-config: opencv
// #cgo LDFLAGS: -L../../../caffe/build/lib -lcaffe -lglog -lboost_system -lboost_thread
// #cgo CXXFLAGS: -std=c++11 -I../../../caffe/include -I.. -O2 -fomit-frame-pointer -Wall
// #include <stdlib.h>
// #include "classification.h"
import "C"
import "unsafe"

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
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
	var job Job
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&job)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "Invalid JSON "+err.Error(), http.StatusBadRequest)
		return
	}

	var modelGob []byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MODELS_BUCKET)
		v := b.Get([]byte(job.Model))
		if v == nil {
			panic("You idiot. You passed in a missing model.")
		}

		modelGob = make([]byte, len(v))

		copy(modelGob, v)
		return nil
	})

	fmt.Println("LEKEKEKEKEKEKE:, ", len(modelGob))

	if err != nil {
		panic("error in ransaction! " + err.Error())
	}

	buf := bytes.NewBuffer(modelGob)
	dec := gob.NewDecoder(buf)

	var model Model
	dec.Decode(&model)
	fmt.Println("HEY %v", model)

	data, err := base64.StdEncoding.DecodeString(job.Image)

	if err != nil {
		panic("Failed to b64 decode image: " + err.Error())
	}

	var cclass C.c_classifier

	cmean := C.CString(model.MeanPath)
	clabel := C.CString(model.LabelsPath)
	cweights := C.CString(model.WeightsPath)
	cmodel := C.CString(model.ModelPath)

	cclass, err = C.classifier_initialize(cmodel, cweights, cmean, clabel)
	if err != nil {
		panic("err in initialize: " + err.Error())
	}

	cstr, err := C.classifier_classify(cclass, (*C.char)(unsafe.Pointer(&data[0])), C.size_t(len(data)))

	if err != nil {
		panic("error classifying: " + err.Error())
	}

	defer C.free(unsafe.Pointer(cstr))

	io.WriteString(w, C.GoString(cstr))

	/*
		job.Output = make(chan string)
		fmt.Println("\nAdded to work queue.")
		WorkQueue <- job

		select {
		case classified := <-job.Output:
			fmt.Println("Request returning.")
			tmp := map[string]string{"output": string(classified)}
			writeResp(w, tmp, 200)
			return
		}*/

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
