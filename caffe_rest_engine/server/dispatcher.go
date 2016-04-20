package main

import (
	"fmt"
)

type Dispatcher struct {
	Policy string
	NWorkers int
	Stop chan bool
	WorkersQueue chan chan Job
}

func NewDispatcher(policy string, nworkers int) Dispatcher {
	dispat := Dispatcher{
		Policy: policy,
		NWorkers: nworkers,
		Stop: make(chan bool),
		WorkersQueue: make(chan chan Job, nworkers)}
	return dispat
}

func (dis Dispatcher) StartDispatcher() {
	for i := 0; i < dis.NWorkers; i++ {
		fmt.Println("Starting worker with id: %d", i)
		worker := NewWorker(i, dis.WorkersQueue)
		worker.Start()
	}

	go func () {
		for {
			select {
			case currJob := <- WorkQueue:
				go func () {
					currWorkerQueue := <- dis.WorkersQueue
					currWorkerQueue <- currJob
					fmt.Printf("Dispatched job with model %s to worker.\n", currJob.Model)
				}()
			case <- dis.Stop:
				fmt.Printf("The dispatcher has been ordered to shutdown. All systems down.\n")
				return
			}
		}
	}()
}

func (dis Dispatcher) Quit() {
	dis.Stop <- true
}


