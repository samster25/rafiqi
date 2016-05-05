package main

import (
	//"fmt"
	//	"math"
	"time"
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
			case <-time.After(time.Duration(QUANTA) * time.Millisecond):
				for el := LRU.Front(); el != nil; el = el.Next() {
					model := (el.Value).(string)
					modelInfo, ok := b.ModelInfo[model]
					if !ok {
						continue
					}
					if modelInfo.count >= modelInfo.threshold && modelInfo.count != 0 {
						modelInfo.threshold = modelInfo.threshold + modelInfo.count
						//fmt.Println("Threshold increased", modelInfo.threshold)
						modelInfo.count = modelInfo.count - MAX_BATCH_AMT
						if modelInfo.count < 0 {
							modelInfo.count = 0
						}
						WorkQueue.batchedJobs <- model

					} else {
						modelInfo.threshold = modelInfo.threshold / 2
						//fmt.Println("Threshold decreased", modelInfo.threshold)
					}
				}
			}
		}
	}()
}
