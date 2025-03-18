package main

// simple consumer; for now just receive and print
// later, use worker pools (goroutines) to handle messages to index algolia
import (
	"log"
	"fmt"
	"context"
	"net/http"
	"os"
	"encoding/json"
	"sync/atomic"

	"github.com/copium-dev/copium/algolia-consumer/inits"
	"github.com/copium-dev/copium/algolia-consumer/job"

	"cloud.google.com/go/pubsub"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
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
// gcloud beta emulators pubsub env-init
// >>>> we need to (1) create `algolia` topic and (2) create a subscription
// >>>> make sure you're logged in to (gcloud auth login)
// gcloud beta emulators pubsub start --project=jtrackerkimpark
// >>>> run the same in ALGOLIA consumer (just change topic name)
func main() {
    // create algolia client (shared across workers)
    algoliaClient, err := inits.InitializeAlgoliaClient()
    if err != nil {
        log.Fatalf("Error initializing algolia client: %v", err)
    }

    // assign IDs to jobs; not exactly necessary but good for tracking and debugging
    var counter int32 = 1

	if os.Getenv("ENVIRONMENT") == "prod" {
		runPushSubscription(algoliaClient, counter)
	} else {
		runPullSubscription(algoliaClient, counter)
	}

}

func runPushSubscription(algoliaClient *search.APIClient, counter int32) {
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

        log.Printf("[*] ALGOLIA [*] Received Pub/Sub message: %s", pubSubMessage.Message.Data)

        // initialize a new job
        jobID := atomic.AddInt32(&counter, 1)
        newJob, err := job.NewJob(pubSubMessage.Message.Data, jobID, algoliaClient)
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

        fmt.Println("Job done, acknowledging message (ALGOLIA)")
        w.WriteHeader(http.StatusOK)
    })

    // Start HTTP server - cloud run will automatically assign PORT variable
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("[*] ALGOLIA [*] Starting push subscription server on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func runPullSubscription(algoliaClient *search.APIClient, counter int32) {
	// create pubsub client and subscription
	sub, pubsubClient, err := inits.InitializeConsumerSubscription()
    if err != nil {
        log.Fatalf("Failed to create Pub/Sub client: %v", err)
    }
    defer pubsubClient.Close()

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
        log.Printf("[*] ALGOLIA [*] Received Pub/Sub message: %s", m.Data)

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

		fmt.Println("Job done, acking message (ALGOLIA)")
		m.Ack()
    })
    if err != nil {
        log.Printf("Error receiving messages: %v", err)
    }

    // block forever (or until process is terminated)
    select {}
}