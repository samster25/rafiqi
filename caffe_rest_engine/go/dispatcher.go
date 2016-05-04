package main

import (
	"fmt"
)

var WorkQueue *HashyLinkedList = NewHashyLinkedList()

type Dispatcher struct {
	Policy       string
	NWorkers     int
	Stop         chan bool
	WorkersQueue chan chan string
}

func NewDispatcher(policy string, nworkers int) Dispatcher {
	dispat := Dispatcher{
		Policy:       policy,
		NWorkers:     nworkers,
		Stop:         make(chan bool),
		WorkersQueue: make(chan chan string, nworkers)}
	return dispat
}

func (dis Dispatcher) StartDispatcher() {
	for i := 0; i < dis.NWorkers; i++ {
		fmt.Println("Starting worker with id: %d", i)
		worker := NewWorker(i, dis.WorkersQueue)
		worker.Start()
	}

	go func() {
		for {
			select {
			case currModel := <-WorkQueue.batchedJobs: //PopFront(MAX_BATCH_AMT)
				Debugf("Current Model %s", currModel)
				go func() {
					currWorkerQueue := <-dis.WorkersQueue
					currWorkerQueue <- currModel
				}()
			}
		}
	}()
	return
}

func (dis Dispatcher) Quit() {
	dis.Stop <- true
}
