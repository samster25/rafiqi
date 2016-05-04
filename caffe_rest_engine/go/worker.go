package main

// #include <stdlib.h>
// #include <classification.h>
import "C"
import "unsafe"
import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"sync"
	"time"
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
	WorkQueue   chan string
	WorkerQueue chan chan string
	Quit        chan bool
}

func handleError(msg string, err error) {
	errorLogger.Printf("%s: %v", msg, err)
}

func NewWorker(id int, workers chan chan string) Worker {
	return Worker{
		ID:          id,
		WorkQueue:   make(chan string),
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
			start := time.Now()
			cclass, err := C.model_init(cmodel, cweights, cmean, clabel)
			fmt.Println("here", m.Name)
			LogTimef("%v model_init", start, m.Name)

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

func (w Worker) classify(job_model string, jobs []Job) []string {
	Debugf(fmt.Sprintf("worker %d beginning classify of %d jobs", w.ID, len(jobs)))
	var modelGob []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MODELS_BUCKET)
		v := b.Get([]byte(job_model))
		if v == nil {
			err := errors.New("Missing model: " + job_model)
			handleError("", err)
			return err
		}

		modelGob = make([]byte, len(v))

		copy(modelGob, v)
		return nil
	})
	if err != nil {
		handleError("error in transaction! ", err)
		return nil
	}

	buf := bytes.NewBuffer(modelGob)
	dec := gob.NewDecoder(buf)
	var model Model
	dec.Decode(&model)

	entry := InitializeModel(&model)
	//entry.Lock()
	start := time.Now()
	batch_mats := make([][]byte, len(jobs))
	for i, job := range jobs {
		batch_mats[i] = job.Image
	}
	lengths := make([]C.size_t, len(jobs))
	for i, el := range batch_mats {
		lengths[i] = C.size_t(len(el))
	}
	cstr_arr, err := C.model_classify_batch(
		entry.Classifier,
		(**C.char)(unsafe.Pointer(&batch_mats[0])),
		(*C.size_t)(unsafe.Pointer(&lengths[0])),
		C.size_t(len(jobs)),
	)
	LogTimef("%v model_classify", start, jobs[0].Model)
	//entry.Unlock()
	//byte_convert := [][]byte(cstr_arr)
	if err != nil {
		handleError("error classifying: ", err)
	}
	go_results := (*[MAX_BATCH_AMT](*C.char))(unsafe.Pointer(cstr_arr))
	final := make([]string, len(jobs))
	for i := 0; i < len(jobs); i++ {
		final[i] = C.GoString(go_results[i])
	}
	//final := make([]string, 1)
	//final[0] = C.GoString(cstr_arr)
	defer C.free(unsafe.Pointer(cstr_arr))
	return final
}

func (w Worker) Start() {
	go func() {
		for {
			w.WorkerQueue <- w.WorkQueue
			select {
			case currModel := <-w.WorkQueue:
				fmt.Println(currModel)
				currJobs := WorkQueue.CreateBatchJob(currModel)
				if len(currJobs) == 0 {
					fmt.Println("There's no jobs for this model")
				}
				if currJobs == nil {
					panic("CURJOBS NIL")
				}
				res := w.classify(currJobs[0].Model, currJobs)
				for i := range res {
					currJobs[i].Output <- res[i]
				}
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

func init() {
	loadedModels = LoadedModelsMap{
		Models: make(map[string]*ModelEntry),
	}

	C.classifier_init()
}
