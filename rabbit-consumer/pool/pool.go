// set up a worker pool to consume and process rabbitmq messages
package pool

import (
	"encoding/json"
	"log"
)

// note: id is not strictly necessary but useful for debugging
type Job struct {
	ID 		  int32
	Data      json.RawMessage
}

// a pool of workers with channels for job distribution, job queueing, and stopping
type Pool struct {
	NumWorkers 	int32
	JobChannels chan chan Job
	JobQueue 	chan Job
	Stopped 	chan bool
}

// a worker with a unique ID and a dedicated job channel (to receive jobs from another goroutine)
// has a shared channel for job registration and a quit channel to signal termination
type Worker struct {
	ID 			int
	JobChannel 	chan Job
	JobChannels chan chan Job
	Quit 		chan bool
}

// initialize a new worker pool 
func NewPool(numWorkers int32) Pool {
	return Pool{
		NumWorkers:  numWorkers,
		JobChannels: make(chan chan Job),
		JobQueue:    make(chan Job),
		Stopped:     make(chan bool),
	}
}

// spawn the worker goroutines and allocates jobs to them
func (p *Pool) Run() {
	log.Println("Spawning the workers")
	for i := 0; i < int(p.NumWorkers); i++ {
		worker := Worker{
			ID:          (i + 1),
			JobChannel:  make(chan Job),
			JobChannels: p.JobChannels,
			Quit:        make(chan bool),
		}
		worker.Start()
	}
	p.Allocate()
}

// pull from the queue and send the job to the channel
func (p *Pool) Allocate() {
	q := p.JobQueue
	s := p.Stopped
	go func(queue chan Job) {
		for {
			select {
			case job := <-q:
				availChannel := <-p.JobChannels
				availChannel <- job

			case <-s:
				return
			}
		}
	}(q)
}

func (w *Worker) Start() {
	log.Printf("Starting Worker ID [%d]", w.ID)
	go func() {
		for {
			w.JobChannels <- w.JobChannel
			select {
			case job := <-w.JobChannel:
				w.work(job)
			case <-w.Quit:
				return
			}

		}
	}()
}

// actually do the job (here is where we want to index algolia)
func (w *Worker) work(job Job) {
	log.Printf("------")
	log.Printf("Processed by Worker [%d]", w.ID)

	// unmarshal the job data
	var data map[string]interface{}
	err := json.Unmarshal(job.Data, &data)
	if err != nil {
		log.Printf("Failed to unmarshal job data: %s", err)
		return
	}

	// process the job (for now just print)
	log.Printf("Job Data: %v", data)

	// 1. figure out the operation (add edit delete)
	// 2. call the correct function w/ data 
	// 3. log the result. done

	// end; log completion
	log.Printf("Processed Job With ID [%d] & content: [%s]", job.ID, job.Data)
	log.Printf("-------")
}