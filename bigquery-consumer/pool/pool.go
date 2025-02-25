// set up a worker pool to consume and process rabbitmq messages
package pool

import (
	"encoding/json"
	"log"
	"fmt"

	"cloud.google.com/go/bigquery"
)

// note: id is not strictly necessary but useful for debugging
type Job struct {
	ID   int32
	Data json.RawMessage
}

// a pool of workers with channels for job distribution, job queueing, and stopping
type Pool struct {
	NumWorkers    int32
	BigQueryClient *bigquery.Client
	JobChannels   chan chan Job
	JobQueue      chan Job
	Stopped       chan bool
}

// a worker with a unique ID and a dedicated job channel (to receive jobs from another goroutine)
// has a shared channel for job registration and a quit channel to signal termination
// note: worker uses shared algolia client
type Worker struct {
	ID            int
	BigQueryClient *bigquery.Client
	JobChannel    chan Job
	JobChannels   chan chan Job
	Quit          chan bool
}

// initialize a new worker pool
func NewPool(numWorkers int32, bigQueryClient *bigquery.Client) Pool {
	return Pool{
		NumWorkers:    numWorkers,
		BigQueryClient: bigQueryClient,
		JobChannels:   make(chan chan Job),
		JobQueue:      make(chan Job),
		Stopped:       make(chan bool),
	}
}

// spawn the worker goroutines and allocates jobs to them
func (p *Pool) Run() {
	log.Println("Spawning the workers (BIGQUERY)")
	for i := 0; i < int(p.NumWorkers); i++ {
		worker := Worker{
			ID:            (i + 1),
			BigQueryClient: p.BigQueryClient,
			JobChannel:    make(chan Job),
			JobChannels:   p.JobChannels,
			Quit:          make(chan bool),
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
    go func() {
        for {
            // not either send w.JobChannel or return if a quit signal is received.
            select {
            case w.JobChannels <- w.JobChannel:
            case <-w.Quit:
                return
            }
            
            select {
            case job := <-w.JobChannel:
                w.work(job)
            case <-w.Quit:
                return
            }
        }
    }()
}

func (w *Worker) work(job Job) {
	log.Printf("[*] BigQuery [*]")
	log.Printf("-------")
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
	operation, ok := data["operation"].(string)
	if !ok {
		log.Printf("Failed to get operation from data")
		return
	}
	log.Printf("Operation: %s", operation)

	fmt.Println(data)

	// end; log completion
	log.Printf("Processed Job With ID [%d] & content: [%s]", job.ID, job.Data)
	log.Printf("-------")
}