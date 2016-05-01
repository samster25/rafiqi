package main

// #include <stdlib.h>
// #include <classification.h>
import "C"
import "unsafe"
import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/boltdb/bolt"
	"sync"
)

type LoadedModelsMap struct {
	sync.RWMutex
	Models map[string]*ModelEntry
}

type ModelEntry struct {
	sync.Mutex
	Classifier C.c_model
}

var loadedModels LoadedModelsMap

type Worker struct {
	ID          int
	WorkQueue   chan []Job
	WorkerQueue chan chan []Job
	Quit        chan bool
}

func handleError(msg string, err error) {
	panic(msg + err.Error())
}

func NewWorker(id int, workers chan chan []Job) Worker {
	return Worker{
		ID:          id,
		WorkQueue:   make(chan []Job),
		WorkerQueue: workers,
		Quit:        make(chan bool)}
}

func InitializeModel(m *Model) *ModelEntry {
	var entry *ModelEntry
	loadedModels.RLock()
	entry, ok := loadedModels.Models[m.Name]
	loadedModels.RUnlock()

	if !ok {
		cmean := C.CString(m.MeanPath)
		clabel := C.CString(m.LabelsPath)
		cweights := C.CString(m.WeightsPath)
		cmodel := C.CString(m.ModelPath)

		loadedModels.Lock()
		// Ensure no one added this model between the RUnlock and here
		_, ok = loadedModels.Models[m.Name]
		if !ok {
			cclass, err := C.model_init(cmodel, cweights, cmean, clabel)

			if err != nil {
				handleError("init failed: ", err)
			}

			entry = &ModelEntry{Classifier: cclass}
			loadedModels.Models[m.Name] = entry
		}
		loadedModels.Unlock()
	}

	return entry

}

func (w Worker) classify(job Job) string {
	fmt.Println("Inside of classify")
	var modelGob []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MODELS_BUCKET)
		fmt.Println(job.Model)
		v := b.Get([]byte(job.Model))
		if v == nil {
			fmt.Println("About to fail: ", job.Model)
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

	//	data, err := base64.StdEncoding.DecodeString(job.Image)

	if err != nil {
		panic("Failed to b64 decode image: " + err.Error())
	}

	entry := InitializeModel(&model)

	if entry.Classifier == nil {
		fmt.Println("fuck me gently with chainsaw")
	}
	entry.Lock()
	cstr, err := C.model_classify(
		entry.Classifier,
		job.Image,
	)
	entry.Unlock()

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
			case currJobs := <-w.WorkQueue:
				fmt.Printf("Job received by a worker (ID: %d)\n", w.ID)
				for _, val := range currJobs {
					val.Output <- w.classify(val)
				}
				//fmt.Printf("The result of classification: %s\n", res)
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

func init() {
	loadedModels = LoadedModelsMap{
		Models: make(map[string]*ModelEntry),
	}

	C.classifier_init()
}
