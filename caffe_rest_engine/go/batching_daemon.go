package main

import (
//	"fmt"
//	"time"
)

type BatchDaemon struct {
	ModelInfo        map[string]*ModelInfoEntry
	QuantaTime       int
	IncrementChannel chan string
}

type ModelInfoEntry struct {
	count          int
	threshold      int
	max_batch_size int
}

func NewModelEntry() *ModelInfoEntry {
	return &ModelInfoEntry{
		count:          0,
		threshold:      0,
		max_batch_size: 32,
	}
}

func NewBatchDaemon() *BatchDaemon {
	bd := &BatchDaemon{
		ModelInfo:        make(map[string]*ModelInfoEntry),
		QuantaTime:       10,
		IncrementChannel: make(chan string),
	}
	return bd
}

func (b *BatchDaemon) Start() {

	for el := LRU.Front(); el != nil; el = el.Next() {
		model := (el.Value).(string)
		b.ModelInfo[model] = NewModelEntry()
	}
	go func() {
		for {
			select {
			case modelString := <-b.IncrementChannel:
				b.ModelInfo[modelString].count++
			default:
				for el := LRU.Front(); el != nil; el = el.Next() {
					model := (el.Value).(string)
					modelInfo := b.ModelInfo[model]
					//if modelInfo.count >= modelInfo.threshold {
					//	modelInfo.count = modelInfo.count - modelInfo.max_batch_size
					//	if modelInfo.count < 0 {
					//		modelInfo.count = 0
					//	}
					if modelInfo.count > 0 {
						WorkQueue.batchedJobs <- model
						modelInfo.count--
					}
				}
			}
		}
	}()
}
