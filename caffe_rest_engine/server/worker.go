package main

import (
	"fmt"
)

type Worker struct {
	ID          int
	WorkQueue   chan Job
	WorkerQueue chan chan Job
	Quit        chan bool
}

func NewWorker(id int, workers chan chan Job) Worker {
	return Worker{
		ID:          id,
		WorkQueue:   make(chan Job),
		WorkerQueue: workers,
		Quit:        make(chan bool)}
}

func (w Worker) classify(job Job) string {
	return job.Model
}

func (w Worker) Start() {
	go func() {
		for {
			w.WorkerQueue <- w.WorkQueue

			select {
			case currJob := <-w.WorkQueue:
				fmt.Printf("Job received by a worker (ID: %d)\n", w.ID)
				res := w.classify(currJob)
				currJob.Output <- res
				fmt.Printf("The result of classification: %s\n", res)
			case <-w.Quit:
				fmt.Println("The worker has been signalled to shut down. Ending now.")
				return
			}
		}
	}()
}

func (w Worker) Stop() {
	go func() {
		w.Quit <- true
	}()
}
