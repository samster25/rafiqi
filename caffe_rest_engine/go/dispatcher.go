package main

import (
	"fmt"
)

var WorkQueue *HashyLinkedList = NewHashyLinkedList()

type Dispatcher struct {
	Policy       string
	NWorkers     int
	Stop         chan bool
	WorkersQueue chan chan []Job
}

func NewDispatcher(policy string, nworkers int) Dispatcher {
	dispat := Dispatcher{
		Policy:       policy,
		NWorkers:     nworkers,
		Stop:         make(chan bool),
		WorkersQueue: make(chan chan []Job, nworkers)}
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
			fmt.Println("wfafhiaf")
			currJobs := WorkQueue.PopFront(MAX_BATCH_AMT)
			fmt.Println(currJobs[0].Model)
			fmt.Println("Hang forever")
			go func() {
				currWorkerQueue := <-dis.WorkersQueue
				currWorkerQueue <- currJobs
				fmt.Printf("Dispatched jobs to worker.\n")
			}()
			return
		}
	}()
}

func (dis Dispatcher) Quit() {
	dis.Stop <- true
}
