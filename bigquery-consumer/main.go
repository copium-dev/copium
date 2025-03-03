package main

// simple consumer; for now just receive and print
// later, use worker pools (goroutines) to handle messages to index algolia
import (
	"log"
	"time"
	"os"
	"fmt"
	"context"
	"sync/atomic"

	"github.com/juhun32/copium/bigquery-consumer/config"
	"github.com/juhun32/copium/bigquery-consumer/pool"

	"cloud.google.com/go/bigquery"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

func initializeBigQueryClient() (*bigquery.Client, error) {
	ctx := context.Background()
	projectID := "jtrackerkimpark" // in prod, retrieve from env vars
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// export PUBSUB_EMULATOR_HOST=localhost:8085
// export PUBSUB_PROJECT_ID=jtrackerkimpark
// gcloud beta emulators pubsub env-init
// >>>> we need to (1) create `bigquery` topic and (2) create a subscription
// >>>> make sure you're logged in to (gcloud auth login)
// gcloud beta emulators pubsub start --project=jtrackerkimpark
// >>>> run the same in algolia consumer (just change topic name)
// or just do run.py lol
func main() {
    // create algolia client (shared across workers)
    bigQueryClient, err := initializeBigQueryClient()
    if err != nil {
        log.Fatalf("Error initializing algolia client: %v", err)
    }

	ctx := context.Background()
	sub, pubsubClient, err := initializeConsumerSubscription()
    if err != nil {
        log.Fatalf("Failed to create Pub/Sub client: %v", err)
    }
    defer pubsubClient.Close()

    // configure worker pool
    cfg := config.NewConfig(10000, bigQueryClient)
    workerPool := pool.NewPool(cfg.NumWorkers, cfg.BigQueryClient)
    workerPool.Run()

    // we'll use a counter to assign IDs to jobs.
    var counter int32 = 1

    // use Pub/Sub's Receive method, which calls the provided callback concurrently.
    // the callback function should acknowledge the message when done.
    err = sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
        log.Printf("Received Pub/Sub message: %s", m.Data)

		// we want to ack the message AFTER processing it, not on enqueue
		// so, we need to create a done channel to signal completion
		// and ensure that we block until the job is done
		// since sub.Receive() itself is concurrent and we use a worker pool,
		// this doesn't block the main thread or other messages
		// NOTE: chan strut{} takes 0 bytes so its the most efficient way for signaling
		// if you're wondering why not ack within the worker, its because documentation
		// states for control flow, the ack should be done in the callback and not in a goroutine
		done := make(chan struct{})

        job := pool.Job{
            ID:   atomic.AddInt32(&counter, 1),
            Data: m.Data,
			Done: done,
        }

        workerPool.JobQueue <- job

		// since done is a chan struct{}, we can simply say 'close(done)' 
		// within the worker goroutine to signal completion and use below to wait
		<-done
		fmt.Println("Job done, acking message (BIGQUERY)")
		m.Ack()
    })
    if err != nil {
        log.Printf("Error receiving messages: %v", err)
    }

    // block forever (or until process is terminated)
    select {}
}

func initializeConsumerSubscription() (*pubsub.Subscription, *pubsub.Client, error) {
    ctx := context.Background()
    projectID := "jtrackerkimpark" // in prod, retrieve from env vars

    var opts []option.ClientOption
    if pubsubEmulatorHost := os.Getenv("PUBSUB_EMULATOR_HOST"); pubsubEmulatorHost != "" {
        log.Printf("Connecting to Pub/Sub emulator at %s", pubsubEmulatorHost)
        opts = append(opts, 
            option.WithEndpoint(pubsubEmulatorHost),
            option.WithoutAuthentication(),
        )
    } else {
        log.Println("PUBSUB_EMULATOR_HOST not set; using credentials")
        opts = append(opts, option.WithCredentialsFile("pubsub-credentials.json"))
    }
    
    client, err := pubsub.NewClient(ctx, projectID, opts...)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to create Pub/Sub client: %w", err)
    }

    subName := "bigquery-sub"
    sub, err := client.CreateSubscription(ctx, subName, pubsub.SubscriptionConfig{
		Topic: client.Topic("applications"),
		AckDeadline: 10 * time.Second,
		EnableMessageOrdering: true,
	})
	if err != nil {
		if err.Error() == "rpc error: code = AlreadyExists desc = Subscription already exists" {
			log.Printf("Subscription already exists, connecting to it")
			sub = client.Subscription(subName)
			return sub, client, nil
		}
		// other error; fail
		client.Close()
		return nil, nil, fmt.Errorf("failed to create subscription: %w", err)
	}
    
    // double check the sub even exists
    exists, err := sub.Exists(ctx)
    if err != nil {
        client.Close()
        return nil, nil, fmt.Errorf("failed to verify subscription existence: %w", err)
    }
    if !exists {
        client.Close()
        return nil, nil, fmt.Errorf("subscription %q does not exist", subName)
    }
    
    return sub, client, nil
}