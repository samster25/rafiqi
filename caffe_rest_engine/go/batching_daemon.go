package main

import (
	//"math"
	"container/list"
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
				info, ok := b.ModelInfo[modelString]
				if ok {
					info.count++
				} else {
					errorLogger.Println("Missing model: ", modelString)
				}
			case <-time.After(time.Duration(QUANTA) * time.Millisecond):
				noJobs := list.New()
				haveJobs := list.New()
				for el := LRU.Front(); el != nil; el = el.Next() {
					model := (el.Value).(string)
					modelInfo, ok := b.ModelInfo[model]
					if !ok {
						continue
					}
					if modelInfo.count >= modelInfo.threshold && modelInfo.count != 0 {
						modelInfo.threshold = modelInfo.threshold + modelInfo.count
						modelInfo.count = modelInfo.count - MAX_BATCH_AMT
						if modelInfo.count < 0 {
							modelInfo.count = 0
						}
						haveJobs.PushBack(model)
						WorkQueue.batchedJobs <- model

					} else {
						modelInfo.threshold = modelInfo.threshold / 2
						noJobs.PushBack(model)
					}
				}
				noJobs.PushBackList(haveJobs)
				LRU = noJobs
			}
		}
	}()
}
