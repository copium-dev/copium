package inits

import (
	"time"
	"os"
	"log"
	"context"
	"strings"
	"fmt"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/pubsub"
	firebase "firebase.google.com/go"
	"cloud.google.com/go/firestore"

	"google.golang.org/api/option"
)

func InitializeBigQueryClient() (*bigquery.Client, error) {
	// use service account credentials, no need to pass in anything
	ctx := context.Background()
	projectID := "jtrackerkimpark" // in prod, retrieve from env vars

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func InitializeFirestoreClient() (*firestore.Client, error) {
    ctx := context.Background()
	    
    conf := &firebase.Config{
        ProjectID: "jtrackerkimpark",
    }       
    
	if firestoreEmulatorHost := os.Getenv("FIRESTORE_EMULATOR_HOST"); firestoreEmulatorHost != "" {
        log.Printf("[*] BIGQUERY [*] Connecting to Firestore emulator at %s", firestoreEmulatorHost)
        conf.DatabaseURL = "http://" + firestoreEmulatorHost
    } else {
        log.Println("[*] BIGQUERY [*] FIRESTORE_EMULATOR_HOST not set; using service account credentials, nothing to pass in")
    }

	app, err := firebase.NewApp(ctx, conf)
    if err != nil {
       return nil, err
    }

    firestoreClient, err := app.Firestore(ctx)
    if err != nil {
        return nil, err
    }

	return firestoreClient, nil
}

func InitializeConsumerSubscription() (*pubsub.Subscription, *pubsub.Client, error) {
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
        log.Println("PUBSUB_EMULATOR_HOST not set; using service account credentials, nothing to pass in")
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
		if strings.Contains(err.Error(), "AlreadyExists") {
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