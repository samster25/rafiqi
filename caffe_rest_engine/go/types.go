package main

import (
	"container/list"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

var LRU *list.List = list.New()

type Model struct {
	Name        string
	WeightsPath string
	ModelPath   string
	LabelsPath  string
	MeanPath    string
}

type ModelBatchEntry struct {
	JobEntries   *list.List
	LRUEntry     *list.Element
	Used         bool
	MaxBatchSize int
}

type HashyLinkedList struct {
	lock        sync.Mutex
	queue       *list.List
	jobs        map[string]*ModelBatchEntry
	lru         *list.List
	batchedJobs chan string
}

func NewModelBatchEntry() *ModelBatchEntry {
	jentry := &ModelBatchEntry{
		JobEntries:   list.New(),
		Used:         false,
		MaxBatchSize: 32,
	}
	return jentry
}

func NewHashyLinkedList() *HashyLinkedList {
	hll := &HashyLinkedList{
		queue:       list.New(),
		jobs:        make(map[string]*ModelBatchEntry),
		lru:         list.New(),
		batchedJobs: make(chan string),
	}
	return hll
}

func (h *HashyLinkedList) AddJob(job Job) {
	h.lock.Lock()
	newElem := h.queue.PushBack(job)
	_, ok := h.jobs[job.Model]
	if !ok {
		h.jobs[job.Model] = NewModelBatchEntry() //list.New()
		//h.jobs[job.Model].LRUEntry = h.lru.PushBack(job.Model)
	}
	currEntry := h.jobs[job.Model]
	currEntry.JobEntries.PushBack(newElem) //Adding to jobs for that model
	currEntry.Used = true
	h.lock.Unlock()
	return
}

func (h *HashyLinkedList) CreateBatchJob(model string) []Job {
	h.lock.Lock()
	modelBatchEntry, ok := h.jobs[model]
	if !ok {
		return nil
	}
	batchAmt := modelBatchEntry.MaxBatchSize
	jobList := modelBatchEntry.JobEntries
	jobListLen := jobList.Len()
	fmt.Println("jl len", jobListLen)

	if jobListLen == 0 {
		return nil
	}

	if batchAmt > jobListLen {
		batchAmt = jobListLen
	}

	result := make([]Job, batchAmt)
	for i := 0; i < batchAmt; i++ {
		currQueuePtr := (jobList.Remove(jobList.Front())).(*list.Element)
		job := (h.queue.Remove(currQueuePtr)).(Job)
		result[i] = job
	}

	h.lock.Unlock()
	return result
}

func NewModelFromURL(name string, modelReq ModelRequest) Model {
	err := os.MkdirAll("../models/"+name, 0755)
	if err != nil {
		panic("Error creating models file: " + err.Error())
	}

	labelsName, err := DownloadAndWrite(name, name+"_labels",
		modelReq.LabelFile.URL, []byte(modelReq.LabelFile.Blob))
	weightsName, err := DownloadAndWrite(name, name+"_weights", modelReq.WeightsFile.URL, []byte(modelReq.WeightsFile.Blob))
	meansName, err := DownloadAndWrite(name, name+"_mean", modelReq.MeanFile.URL, []byte(modelReq.MeanFile.Blob))
	modelName, err := DownloadAndWrite(name, name+"_mod", modelReq.ModFile.URL, []byte(modelReq.ModFile.Blob))

	return Model{
		Name:        name,
		WeightsPath: weightsName,
		ModelPath:   modelName,
		LabelsPath:  labelsName,
		MeanPath:    meansName,
	}
}

func DownloadAndWrite(dirname string, filename string, url string, blob []byte) (string, error) {
	if url == "" && len(blob) == 0 {
		return "", nil
	}

	fname := fmt.Sprintf("../models/%s/%s", dirname, filename)
	out, err := os.Create(fname)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if len(blob) == 0 {

		// Get the data
		resp, err := http.Get(url)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		// Writer the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return "", err
		}
	} else {
		data, err := base64.StdEncoding.DecodeString(string(blob))
		if err != nil {
			panic("failed to b64decode: " + err.Error())
		}
		ioutil.WriteFile(fname, data, 0755)
	}

	return fname, nil
}
