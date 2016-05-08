package main

import (
	"fmt"
	"math"
	"time"
)

type BatchDaemon struct {
	ModelInfo        *ModelInfoEntry
	QuantaTime       int
	IncrementChannel chan string
	Model            string
	quantaChan       chan int
}

type ModelInfoEntry struct {
	count     int
	threshold float64
}

func NewModelEntry() *ModelInfoEntry {
	return &ModelInfoEntry{
		count:     0,
		threshold: 0,
	}
}

func NewBatchDaemon(model string) *BatchDaemon {
	bd := &BatchDaemon{
		ModelInfo:        NewModelEntry(),
		QuantaTime:       10,
		IncrementChannel: make(chan string),
		Model:            model,
		quantaChan:       make(chan int),
	}
	return bd
}

func (b *BatchDaemon) Start() {

	go func() {
		//waitChan := time.After(time.Duration(QUANTA) * time.Millisecond)
		time_sum := 0
		time_cnt := 0
		avg := QUANTA
		for {
			select {
			case <-b.IncrementChannel:
				b.ModelInfo.count++
			case wait := <-b.quantaChan:
				time_sum = time_sum + wait
				time_cnt++
				avg = int64(math.Ceil(float64(time_sum) / float64(time_cnt)))
				//waitChan = time.After(time.Duration(avg) * time.Millisecond)
				fmt.Println(avg)
			case <-time.After(time.Duration(avg) * time.Millisecond):
				model := b.Model
				modelInfo := b.ModelInfo
				if modelInfo.count >= int(math.Ceil(modelInfo.threshold)) && modelInfo.count != 0 {
					modelInfo.threshold = (1.0-ALPHA)*float64(modelInfo.threshold) + ALPHA*float64(modelInfo.count) //modelInfo.threshold + modelInfo.count
					modelInfo.count = modelInfo.count - MAX_BATCH_AMT
					if modelInfo.count < 0 {
						modelInfo.count = 0
					}
					WorkQueue.batchedJobs <- model
				} else {
					modelInfo.threshold = modelInfo.threshold / 2
					if modelInfo.threshold < 1 {
						modelInfo.threshold = 1
					}
				}
				//noJobs.PushBackList(haveJobs)
				//LRU = noJobs
			}
		}
	}()
}
