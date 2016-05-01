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

type Model struct {
	Name        string
	WeightsPath string
	ModelPath   string
	LabelsPath  string
	MeanPath    string
}

type HashyLinkedList struct {
	lock    sync.Mutex
	condVar *sync.Cond
	queue   *list.List
	jobs    map[string]*list.List
}

func NewHashyLinkedList() *HashyLinkedList {
	hll := &HashyLinkedList{
		queue: list.New(),
		jobs:  make(map[string]*list.List),
	}
	hll.condVar = sync.NewCond(&hll.lock)
	return hll
}

func (h *HashyLinkedList) AddJob(job Job) {
	h.lock.Lock()
	newElem := h.queue.PushBack(job)
	_, ok := h.jobs[job.Model]
	if !ok {
		h.jobs[job.Model] = list.New()
	}
	h.jobs[job.Model].PushBack(newElem)
	h.condVar.Signal()
	h.lock.Unlock()
	return
}

func (h *HashyLinkedList) PopFront(batchAmt int) []Job {
	h.lock.Lock()
	defer h.lock.Unlock()
	for h.queue.Len() == 0 {
		h.condVar.Wait()
	}
	frontElem := h.queue.Front()
	if frontElem == nil {
		return nil
	}
	frontJob := (frontElem.Value).(Job)
	jobList, ok := h.jobs[frontJob.Model]
	if !ok {
		return nil
	}
	jobListLen := jobList.Len()

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
