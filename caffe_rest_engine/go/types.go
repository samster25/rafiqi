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
	ModelSize   int64
}

func (m *Model) estimatedGPUMemSize() uint64 {
	return uint64(K_CONTEXTS * (m.ModelSize + FRAME_BUF_SIZE))
}

type HashyLinkedList struct {
	lock sync.Mutex
	//	queue       *list.List
	jobs        map[string]*HLLEntry
	batchedJobs chan string
}

type HLLEntry struct {
	lock    sync.Mutex
	jobList *list.List
}

func NewHashyLinkedList() *HashyLinkedList {
	hll := &HashyLinkedList{
		//		queue: list.New(),
		jobs:        make(map[string]*HLLEntry),
		batchedJobs: make(chan string),
	}
	return hll
}

func NewHLLEntry() *HLLEntry {
	return &HLLEntry{
		jobList: list.New(),
	}
}

func (h *HashyLinkedList) AddJob(job Job) {
	//newElem := h.queue.PushBack(job)
	hllEntry, ok := h.jobs[job.Model]
	if !ok {
		fmt.Println("Model not found")
	}
	hllEntry.lock.Lock()
	hllEntry.jobList.PushBack(job)
	hllEntry.lock.Unlock() //Adding to jobs for that model
	return
}

func (h *HashyLinkedList) CreateBatchJob(model string) []Job {
	hllEntry, ok := h.jobs[model]
	if !ok {
		return nil
	}

	hllEntry.lock.Lock()
	defer hllEntry.lock.Unlock()
	batchAmt := MAX_BATCH_AMT
	jobList := hllEntry.jobList
	jobListLen := jobList.Len()

	if jobListLen == 0 {
		return nil
	}

	if batchAmt > jobListLen {
		batchAmt = jobListLen
	}

	result := make([]Job, batchAmt)
	for i := 0; i < batchAmt; i++ {
		job := (jobList.Remove(jobList.Front())).(Job)
		//job := (h.queue.Remove(currQueuePtr)).(Job)
		result[i] = job
	}
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

	modelFile, err := os.Open(weightsName)
	if err != nil {
		panic("Error opening model: " + err.Error())
	}
	info, err := modelFile.Stat()
	if err != nil {
		panic("Error stat'ing model: " + err.Error())
	}

	modelSize := info.Size()
	Debugf("Set model size to: %d", modelSize)

	return Model{
		Name:        name,
		WeightsPath: weightsName,
		ModelPath:   modelName,
		LabelsPath:  labelsName,
		MeanPath:    meansName,
		ModelSize:   modelSize,
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
