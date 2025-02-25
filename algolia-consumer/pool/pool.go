// set up a worker pool to consume and process rabbitmq messages
package pool

import (
	"encoding/json"
	"log"
	"fmt"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
)

// note: id is not strictly necessary but useful for debugging
type Job struct {
	ID   int32
	Data json.RawMessage
}

// a pool of workers with channels for job distribution, job queueing, and stopping
type Pool struct {
	NumWorkers    int32
	AlgoliaClient *search.APIClient
	JobChannels   chan chan Job
	JobQueue      chan Job
	Stopped       chan bool
}

// a worker with a unique ID and a dedicated job channel (to receive jobs from another goroutine)
// has a shared channel for job registration and a quit channel to signal termination
// note: worker uses shared algolia client
type Worker struct {
	ID            int
	AlgoliaClient *search.APIClient
	JobChannel    chan Job
	JobChannels   chan chan Job
	Quit          chan bool
}

// initialize a new worker pool
func NewPool(numWorkers int32, algoliaClient *search.APIClient) Pool {
	return Pool{
		NumWorkers:    numWorkers,
		AlgoliaClient: algoliaClient,
		JobChannels:   make(chan chan Job),
		JobQueue:      make(chan Job),
		Stopped:       make(chan bool),
	}
}

// spawn the worker goroutines and allocates jobs to them
func (p *Pool) Run() {
	log.Println("Spawning the workers (ALGOLIA)")
	for i := 0; i < int(p.NumWorkers); i++ {
		worker := Worker{
			ID:            (i + 1),
			AlgoliaClient: p.AlgoliaClient,
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

// actually do the job (here is where we want to index algolia)
func (w *Worker) work(job Job) {
	log.Printf("[*] Algolia [*]")
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

	// we know the operation now, we can delete it.
	delete(data, "operation")

	// 2. call the correct function w/ data
	if operation == "add" {
		w.addApplication(data)
	} else if operation == "edit" {
		w.editApplication(data)
	} else if operation == "delete" {
		w.deleteApplication(data)
	} else if operation == "userDelete" {
		w.userDelete(data)
	} else {
		log.Printf("Unknown operation: %s", operation)
	}

	// end; log completion
	log.Printf("Processed Job With ID [%d] & content: [%s]", job.ID, job.Data)
	log.Printf("-------")
}

func (w *Worker) addApplication(data map[string]interface{}) {
	// add the application to algolia
	saveRes, err := w.AlgoliaClient.SaveObject(
		w.AlgoliaClient.NewApiSaveObjectRequest("users", data),
	)
	if err != nil {
		log.Printf("Failed to save object: %s", err)
		return
	}

	// wait for task to finish before exiting function
	_, err = w.AlgoliaClient.WaitForTask("users", saveRes.TaskID)
	if err != nil {
		log.Printf("Error waiting for task to finish: %s", err)
		return
	}

	log.Printf("Saved object: %v", saveRes)
}

func (w *Worker) editApplication(data map[string]interface{}) {
	// edit the application in algolia
	objectID, ok := data["objectID"].(string)
	if !ok {
		log.Printf("Failed to get objectID from data")
		return
	}

	updateRes, err := w.AlgoliaClient.PartialUpdateObject(
		w.AlgoliaClient.NewApiPartialUpdateObjectRequest("users", objectID, data),
	)
	if err != nil {
		log.Printf("Failed to update object: %s", err)
		return
	}

	// wait for task to finish before exiting function
	_, err = w.AlgoliaClient.WaitForTask("users", *updateRes.TaskID)
	if err != nil {
		log.Printf("Error waiting for task to finish: %s", err)
		return
	}

	log.Printf("Updated object: %v", updateRes)
}

func (w *Worker) deleteApplication(data map[string]interface{}) {
	// delete the application from algolia
	objectID, ok := data["objectID"].(string)
	if !ok {
		log.Printf("Failed to get objectID from data")
		return
	}

	deleteRes, err := w.AlgoliaClient.DeleteObject(
		w.AlgoliaClient.NewApiDeleteObjectRequest("users", objectID),
	)
	if err != nil {
		log.Printf("Failed to delete object: %s", err)
		return
	}

	// wait for task to finish before exiting function
	_, err = w.AlgoliaClient.WaitForTask("users", deleteRes.TaskID)
	if err != nil {
		log.Printf("Error waiting for task to finish: %s", err)
		return
	}
	
	log.Printf("Deleted object: %v", deleteRes)
}

// note: DeleteBy is resource intensive so we should carefully monitor
func (w *Worker) userDelete(data map[string]interface{}) {
	// extract and delete every objectID where email == data["email"]
	email, ok := data["email"].(string)
	if !ok {
		log.Printf("Failed to get email from data")
		return
	}
	
	filter := fmt.Sprintf("email:%s", email)

	res, err := w.AlgoliaClient.DeleteBy(
		w.AlgoliaClient.NewApiDeleteByRequest(
			"users",
			search.NewEmptyDeleteByParams().SetFilters(filter),
		),
	)
	if err != nil {
		log.Printf("Failed to delete by: %s", err)
		return
	}

	// wait for task to finish before exiting function
	_, err = w.AlgoliaClient.WaitForTask("users", res.TaskID)
	if err != nil {
		log.Printf("Error waiting for task to finish: %s", err)
		return
	}

	log.Printf("Deleted objects: %s", res)
}