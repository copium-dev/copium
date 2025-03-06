package inits

import (
	"time"
	"os"
	"log"
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/pubsub"
	firebase "firebase.google.com/go"
	"cloud.google.com/go/firestore"

	"google.golang.org/api/option"
)

func InitializeBigQueryClient() (*bigquery.Client, error) {
	// we can reuse pubsub credentials since they're both in the same Google Cloud project
	// note that Firebase != Google Cloud, so we can't reuse Firebase credentials
	// there's also no emulator for this so we gotta directly go to prod lol
	opt := option.WithCredentialsFile("pubsub-credentials.json")
	ctx := context.Background()
	projectID := "jtrackerkimpark" // in prod, retrieve from env vars

	client, err := bigquery.NewClient(ctx, projectID, opt)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func InitializeFirestoreClient() (*firestore.Client, error) {
	opt := option.WithCredentialsFile("jtracker-backend-credentials.json")
    ctx := context.Background()
	    
    conf := &firebase.Config{
        ProjectID: "jtrackerkimpark-90318",
    }       
    
	if firestoreEmulatorHost := os.Getenv("FIRESTORE_EMULATOR_HOST"); firestoreEmulatorHost != "" {
        log.Printf("[*] BIGQUERY [*] Connecting to Firestore emulator at %s", firestoreEmulatorHost)
        conf.DatabaseURL = "http://" + firestoreEmulatorHost
    } else {
        log.Println("[*] BIGQUERY [*] FIRESTORE_EMULATOR_HOST not set")
    }

	app, err := firebase.NewApp(ctx, conf, opt)
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
