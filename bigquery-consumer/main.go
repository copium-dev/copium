package main

// simple consumer; for now just receive and print
// later, use worker pools (goroutines) to handle messages to index algolia
import (
	"context"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/copium-dev/copium/bigquery-consumer/inits"
	"github.com/copium-dev/copium/bigquery-consumer/job"

	"cloud.google.com/go/pubsub"
)

// export PUBSUB_EMULATOR_HOST=localhost:8085
// export PUBSUB_PROJECT_ID=jtrackerkimpark
// export FIRESTORE_EMULATOR_HOST=localhost:8080
// gcloud beta emulators pubsub env-init
// >>>> we need to (1) create `bigquery` topic and (2) create a subscription
// >>>> make sure you're logged in to (gcloud auth login)
// gcloud beta emulators pubsub start --project=jtrackerkimpark
// >>>> run the same in algolia consumer (just change topic name)
// or just do run.py lol
func main() {
    // create bigquery client (shared across workers)
    bigQueryClient, err := inits.InitializeBigQueryClient()
    if err != nil {
        log.Fatalf("Error initializing algolia client: %v", err)
    }

	// create firestore client (shared across workers)
	firestoreClient, err := inits.InitializeFirestoreClient()
	if err != nil {
		log.Fatalf("Error initializing firestore client: %v", err)
	}
	defer firestoreClient.Close()

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
        log.Printf("[*] BIGQUERY [*] Received Pub/Sub message: %s", m.Data)

		jobID := atomic.AddInt32(&counter, 1)

		// create a new job with necessary data received from pubsub
		newJob, err := job.NewJob(m.Data, jobID, bigQueryClient, firestoreClient)
		if err != nil {
			log.Printf("Failed to create job %d: %s", jobID, err)
			return
		}

		// process the job, use the same context as the aprent
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