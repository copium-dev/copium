package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"os"
	"net/http"
	"encoding/json"
	
	"github.com/copium-dev/copium/bigquery-consumer/inits"
	"github.com/copium-dev/copium/bigquery-consumer/job"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
)

type PubSubMessage struct {
    Message struct {
        Data []byte `json:"data,omitempty"`
        ID   string `json:"id"`
        Attributes map[string]string `json:"attributes,omitempty"`
    } `json:"message"`
    Subscription string `json:"subscription"`
}

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
        log.Fatalf("Error initializing BigQuery client: %v", err)
    }

	// create firestore client (shared across workers)
	firestoreClient, err := inits.InitializeFirestoreClient()
	if err != nil {
		log.Fatalf("Error initializing Firestore client: %v", err)
	}
	defer firestoreClient.Close()

    // assign IDs to jobs; not exactly necessary but good for tracking and debugging
    var counter int32 = 1

	if os.Getenv("ENVIRONMENT") == "prod" {
		runPushSubscription(bigQueryClient, firestoreClient, counter)
	} else {
		runPullSubscription(bigQueryClient, firestoreClient, counter)
	}

}

// runPushSubscription starts the HTTP server for push-based subscription
// return 2XX for ack, 4xx for non-retryable error, 5xx for retryable error
func runPushSubscription(bigQueryClient *bigquery.Client, firestoreClient *firestore.Client, counter int32) {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // only allow POST requests
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

		// parse pubsub message
        var pubSubMessage PubSubMessage
        if err := json.NewDecoder(r.Body).Decode(&pubSubMessage); err != nil {
            log.Printf("Error parsing Pub/Sub message: %v", err)
            http.Error(w, fmt.Sprintf("Error parsing message: %v", err), http.StatusBadRequest)
            return
        }

        log.Printf("[*] BIGQUERY [*] Received Pub/Sub message: %s", pubSubMessage.Message.Data)

        // initialize a new job
        jobID := atomic.AddInt32(&counter, 1)
        newJob, err := job.NewJob(pubSubMessage.Message.Data, jobID, bigQueryClient, firestoreClient)
        if err != nil {
            log.Printf("Failed to create job %d: %s", jobID, err)
			// non-retryable error because it is related to incorrect message format
            http.Error(w, fmt.Sprintf("Failed to create job: %v", err), http.StatusBadRequest)
            return
        }

		// execute job
        ctx := context.Background()
        err = newJob.Process(ctx)
        if err != nil {
            log.Printf("Failed to process job %d: %s", jobID, err)
            http.Error(w, fmt.Sprintf("Failed to process job: %v", err), http.StatusInternalServerError)
            return
        }

        fmt.Println("Job done, acknowledging message (BIGQUERY)")
        w.WriteHeader(http.StatusOK)
    })

    // Start HTTP server - cloud run will automatically assign PORT variable
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("[*] BIGQUERY [*] Starting push subscription server on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func runPullSubscription(bigQueryClient *bigquery.Client, firestoreClient *firestore.Client, counter int32) {
	// create pubsub client and subscription
	sub, pubsubClient, err := inits.InitializeConsumerSubscription()
    if err != nil {
        log.Fatalf("Failed to create Pub/Sub client: %v", err)
    }
    defer pubsubClient.Close()// limit max number of msgs we can receive at once

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

		// process the job, use the same context as the parent
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