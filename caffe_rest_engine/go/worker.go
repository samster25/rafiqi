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
	WorkQueue   chan []Job
	WorkerQueue chan chan []Job
	Quit        chan bool
}

func handleError(msg string, err error) {
	errorLogger.Printf("%s: %v", msg, err)
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
	entry.Lock()
	start := time.Now()
	batch_mats := make([]C.c_mat, len(jobs))
	for i, job := range jobs {
		batch_mats[i] = job.Image
	}
	cstr_arr, err := C.model_classify_batch(
		entry.Classifier,
		(*C.c_mat)(unsafe.Pointer(&batch_mats[0])),
		C.int(len(batch_mats)),
	)
	LogTimef("%v model_classify", start, jobs[0].Model)
	entry.Unlock()
	//byte_convert := [][]byte(cstr_arr)
	if err != nil {
		handleError("error classifying: ", err)
	}
	go_results := (*[MAX_BATCH_AMT](*C.char))(unsafe.Pointer(cstr_arr))
	final := make([]string, len(jobs))
	for i := 0; i < len(jobs); i++ {
		final[i] = C.GoString(go_results[i])
	}
	defer C.free(unsafe.Pointer(cstr_arr))
	return final
}

func (w Worker) Start() {
	go func() {
		for {
			w.WorkerQueue <- w.WorkQueue
            var jobs []Job
			select {
			case currJobs := <-w.WorkQueue:
			    jobs = currJobs	
                //if len(res) == 0 {
				//	res = "Error in classify. see error log for details."
				//}
				//fmt.Printf("The result of classification: %s\n", res)
			case <-w.Quit:
				fmt.Println("The worker has been signalled to shut down. Ending now.")
				return
			}
		    res := w.classify(jobs[0].Model, jobs)
            for i := range res {
                jobs[i].Output <- res[i]
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
