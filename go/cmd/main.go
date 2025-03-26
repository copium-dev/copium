package main

import (
    "log"
    "context"
	"strings"
	"fmt"
    "os"

    "github.com/copium-dev/copium/go/cmd/api"
    "github.com/copium-dev/copium/go/utils"

	"cloud.google.com/go/pubsub"
	firebase "firebase.google.com/go"
	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
    "google.golang.org/api/option"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
)

func main() {
    // initialize firestore; use service account credentials so nothing to do
	firestoreClient, err := initializeFirestoreClient()
	if err != nil {
		log.Fatal("Failed to initialize Firestore client: ", err)
	}
	defer firestoreClient.Close()

	// initialize auth handler
    authHandler := utils.NewAuthHandler()

	// initialize bigquery client
	bigQueryClient, err := initializeBigQueryClient()
	if err != nil {
		log.Fatal("Failed to initialize BigQuery client: ", err)
	}

	// initialize pubsub (two topics: `algolia` and `bigquery`)
	// typically will be port 8085 
	pubsubClient, applicationsTopic, err := initializePubSubClient()
	if err != nil {
		log.Fatal("Failed to initialize Pub/Sub client: ", err)
	}

	// after server stops, clean up pub/sub topic goroutines
	// if you are confused why, here is a snippet from https://pkg.go.dev/cloud.google.com/go/pubsub#section-readme
	// >>>> The first time you call Topic.Publish on a Topic, goroutines are started
	// >>>> in the background. To clean up these goroutines, call Topic.Stop
	defer applicationsTopic.Stop()
	defer pubsubClient.Close()

	// initialize algolia client (read-only)
	algoliaClient, err := initializeAlgoliaClient()
	if err != nil {
		log.Fatal("Failed to initialize Algolia client: ", err)
	}

	pubSubOrderingKey := os.Getenv("PUBSUB_ORDERING_KEY")


	// cloud run will provide PORT 8080 by default in env
    port := os.Getenv("PORT")
    if port == "" {
        port = "8000"
    }

    log.Printf("Starting server on port %s", port)

	server := api.NewAPIServer(":" + port, firestoreClient, algoliaClient, bigQueryClient, authHandler, applicationsTopic, pubSubOrderingKey)
    if err := server.Run(); err != nil {
        log.Fatal(err)
    }
}

func initializeAlgoliaClient() (*search.APIClient, error) {
	appID := os.Getenv("ALGOLIA_APP_ID")
	searchApiKey := os.Getenv("ALGOLIA_SEARCH_API_KEY")

	algoliaClient, err := search.NewClient(appID, searchApiKey)
	if err != nil {
		return nil, err
	}

	return algoliaClient, nil
}

func initializePubSubClient() (*pubsub.Client, *pubsub.Topic, error) {
    ctx := context.Background()
    projectID := "jtrackerkimpark" // in prod, use env vars

	var opts []option.ClientOption

	// if PUBSUB_EMULATOR_HOST is set, use it; otherwise use credentials file
	// if credentials file is used WE ARE WORKING IN PROD so be careful
    if pubsubEmulatorHost := os.Getenv("PUBSUB_EMULATOR_HOST"); pubsubEmulatorHost != "" {
        log.Printf("Connecting to Pub/Sub emulator at %s", pubsubEmulatorHost)
        // use both the endpoint option and disable authentication.
        opts = append(opts, 
            option.WithEndpoint(pubsubEmulatorHost),
            option.WithoutAuthentication(),
        )
    } else {
        log.Println("PUBSUB_EMULATOR_HOST not set; using service account credentials, nothing to pass in")
    }

    pubsubClient, err := pubsub.NewClient(ctx, projectID)
    if err != nil {
        return nil, nil, err
    }

	var applicationsTopic *pubsub.Topic

	// create topic (algolia and bigquery both subscribe to this topic)
	applicationsTopic, err = pubsubClient.CreateTopic(ctx, "applications")
	if err != nil {
		if strings.Contains(err.Error(), "AlreadyExists") {
			applicationsTopic = pubsubClient.Topic("applications")
			fmt.Println("applications topic already exists, connecting to it")
		} else {
			return nil, nil,  err
		}
	}

	applicationsTopic.EnableMessageOrdering = true

    return pubsubClient, applicationsTopic, nil
}

func initializeBigQueryClient() (*bigquery.Client, error) {
	// use service account credentials, no need to pass in anything
	ctx := context.Background()
	projectID := "jtrackerkimpark" // in prod, retrieve from env vars

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func initializeFirestoreClient() (*firestore.Client, error) {
    ctx := context.Background()
	    
    conf := &firebase.Config{
        ProjectID: "jtrackerkimpark",
    }       
    
	if firestoreEmulatorHost := os.Getenv("FIRESTORE_EMULATOR_HOST"); firestoreEmulatorHost != "" {
        log.Printf("Connecting to Firestore emulator at %s", firestoreEmulatorHost)
        conf.DatabaseURL = "http://" + firestoreEmulatorHost
    } else {
        log.Println("FIRESTORE_EMULATOR_HOST not set; using service account credentials, nothing to pass in")
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
