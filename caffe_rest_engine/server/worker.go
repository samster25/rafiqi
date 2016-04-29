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
	"fmt"

	"github.com/boltdb/bolt"
)

type Worker struct {
	ID          int
	WorkQueue   chan Job
	WorkerQueue chan chan Job
	Quit        chan bool
}

func NewWorker(id int, workers chan chan Job) Worker {
	return Worker{
		ID:          id,
		WorkQueue:   make(chan Job),
		WorkerQueue: workers,
		Quit:        make(chan bool)}
}

func (w Worker) classify(job Job) string {
	var modelGob []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MODELS_BUCKET)
		v := b.Get([]byte(job.Model))
		if v == nil {
			panic("You idiot. You passed in a missing model.")
		}

		modelGob = make([]byte, len(v))

		copy(modelGob, v)
		return nil
	})

	if err != nil {
		panic("error in transaction! " + err.Error())
	}

	buf := bytes.NewBuffer(modelGob)
	dec := gob.NewDecoder(buf)

	var model Model
	dec.Decode(&model)

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
	defer C.free(unsafe.Pointer(cclass))
	if err != nil {
		panic("err in initialize: " + err.Error())
	}

	cstr, err := C.classifier_classify(cclass, (*C.char)(unsafe.Pointer(&data[0])), C.size_t(len(data)))

	if err != nil {
		panic("error classifying: " + err.Error())
	}

	defer C.free(unsafe.Pointer(cstr))
	return C.GoString(cstr)
}

func (w Worker) Start() {
	go func() {
		for {
			w.WorkerQueue <- w.WorkQueue

			select {
			case currJob := <-w.WorkQueue:
				fmt.Printf("Job received by a worker (ID: %d)\n", w.ID)
				res := w.classify(currJob)
				currJob.Output <- res
				fmt.Printf("The result of classification: %s\n", res)
			case <-w.Quit:
				fmt.Println("The worker has been signalled to shut down. Ending now.")
				return
			}
		}
	}()
}

func (w Worker) Stop() {
	go func() {
		w.Quit <- true
	}()
}
