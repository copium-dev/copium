package main

// simple consumer; for now just receive and print
// later, use worker pools (goroutines) to handle messages to index algolia
import (
	"log"
	"fmt"
	"context"
	"sync/atomic"

	"github.com/copium-dev/copium/algolia-consumer/inits"
	"github.com/copium-dev/copium/algolia-consumer/job"

	"cloud.google.com/go/pubsub"
)

// export PUBSUB_EMULATOR_HOST=localhost:8085
// export PUBSUB_PROJECT_ID=jtrackerkimpark
// gcloud beta emulators pubsub env-init
// >>>> we need to (1) create `algolia` topic and (2) create a subscription
// >>>> make sure you're logged in to (gcloud auth login)
// gcloud beta emulators pubsub start --project=jtrackerkimpark
// >>>> run the same in bigquery consumer (just change topic name)
func main() {
    // create algolia client (shared across workers)
    algoliaClient, err := inits.InitializeAlgoliaClient()
    if err != nil {
        log.Fatalf("Error initializing algolia client: %v", err)
    }

	// create pubsub client and subscription
	sub, pubsubClient, err := inits.InitializeConsumerSubscription()
    if err != nil {
        log.Fatalf("Failed to create Pub/Sub client: %v", err)
    }
    defer pubsubClient.Close()

    // assign IDs to jobs; not exactly necessary but good for tracking and debugging
    var counter int32 = 1
	
	// limit max number of msgs we can receive at once
	sub.ReceiveSettings.MaxOutstandingMessages = 1000
	// limit max number of goroutines spawned to process messages
	sub.ReceiveSettings.NumGoroutines = 100

	ctx := context.Background()

	// NOTE: previously we were using our own worker pool (because of RabbitMQ) but it makes no sense to when
	// 		 sub.Receive handles concurrent message handling for us 
    // use Pub/Sub's Receive method, which calls the provided callback asynchronously.
	// ack is only called when message is successfully processed; otherwise message is redelivered
    err = sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
        log.Printf("Received Pub/Sub message: %s", m.Data)

		jobID := atomic.AddInt32(&counter, 1)

		newJob, err := job.NewJob(m.Data, jobID, algoliaClient)
		if err != nil {
			log.Printf("Failed to create job %d: %s", jobID, err)
			return
		}

		err = newJob.Process(ctx)
		if err != nil {
			log.Printf("Failed to process job %d: %s", jobID, err)
			return
		}

		fmt.Println("Job done, acking message (BIGQUERY)")
		m.Ack()
    })
    if err != nil {
        log.Printf("Error receiving messages: %v", err)
    }

    // block forever (or until process is terminated)
    select {}
}