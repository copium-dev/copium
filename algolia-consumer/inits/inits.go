package inits

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func InitializeAlgoliaClient() (*search.APIClient, error) {
	if os.Getenv("ENVIRONMENT") != "prod" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file")
		}
	}

	appID := os.Getenv("ALGOLIA_APP_ID")
	writeApiKey := os.Getenv("ALGOLIA_WRITE_API_KEY")

	algoliaClient, err := search.NewClient(appID, writeApiKey)
	if err != nil {
		return nil, err
	}

	return algoliaClient, nil
}

func InitializeConsumerSubscription() (*pubsub.Subscription, *pubsub.Client, error) {
	ctx := context.Background()
	projectID := "jtrackerkimpark" // in prod, retrieve from env vars

	// configure whether to be in prod or emulator
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

	// establish connection, uses opts above to determine whether to use emulator or credentials (for prod)
	client, err := pubsub.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Pub/Sub client: %w", err)
	}

	// create subscription to the `applications` topic (if it doesnt exist)
	subName := "algolia-sub"
	sub, err := client.CreateSubscription(ctx, subName, pubsub.SubscriptionConfig{
		Topic:                 client.Topic("applications"),
		AckDeadline:           10 * time.Second,
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