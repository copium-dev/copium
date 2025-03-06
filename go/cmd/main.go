package main

import (
    "log"
    "context"
	"fmt"
    "os"

    "github.com/copium-dev/copium/go/cmd/api"
    "github.com/copium-dev/copium/go/utils"

	"cloud.google.com/go/pubsub"

    firebase "firebase.google.com/go"
    "google.golang.org/api/option"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
)

func main() {
    // initialize firestore; requires a credentials json file 
    opt := option.WithCredentialsFile("jtracker-backend-credentials.json")
    ctx := context.Background()
    
    conf := &firebase.Config{
        ProjectID: "jtrackerkimpark-90318",
    }       
    
    // create firestore emulator connection (in prod we can literally just delete this if statement and it
    // will connect to prod firestore instance)
    // BEFORE RUNNING `go run cmd/main.go`:
    // 1. cd into `go/`
    // 2. run `firebase emulators:start` (starts firestore emulator on localhost:8080)
    // 3. open firestore emulator UI and add `users` collection
    // 4. open another terminal and cd into `go/`
    // 5. run `export FIRESTORE_EMULATOR_HOST=localhost:8080`
    //    - if windows, run `$env:FIRESTORE_EMULATOR_HOST="localhost:8080"`
    // 6. run `go run cmd/main.go`
    if firestoreEmulatorHost := os.Getenv("FIRESTORE_EMULATOR_HOST"); firestoreEmulatorHost != "" {
        log.Printf("Connecting to Firestore emulator at %s", firestoreEmulatorHost)
        conf.DatabaseURL = "http://" + firestoreEmulatorHost
    } else {
        log.Println("FIRESTORE_EMULATOR_HOST not set")
    }

    app, err := firebase.NewApp(ctx, conf, opt)
    if err != nil {
        log.Fatal(err)
    }

    firestoreClient, err := app.Firestore(ctx)
    if err != nil {
        log.Fatal(err)
    }

    defer firestoreClient.Close()

    port := os.Getenv("PORT")
    if port == "" {
        port = "8000"
    }

    log.Printf("Starting server on port %s", port)

    authHandler := utils.NewAuthHandler()

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

	pubSubOrderingKey := "a_random_key_haha"

    // temp: firestore emulator is on 8080 so use 8000 for API server
    // in prod, use Google Cloud Run's default PORT env variable
    // In main.go, modify your server initialization:
	server := api.NewAPIServer(":" + port, firestoreClient, algoliaClient, authHandler, applicationsTopic, pubSubOrderingKey)
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

	// if PUBSUB_EMULATOR_HOST is set, use it; otherwise use credentials file
	// if credentials file is used WE ARE WORKING IN PROD so be careful
    var opts []option.ClientOption
    if pubsubEmulatorHost := os.Getenv("PUBSUB_EMULATOR_HOST"); pubsubEmulatorHost != "" {
        log.Printf("Connecting to Pub/Sub emulator at %s", pubsubEmulatorHost)
        // Use both the endpoint option and disable authentication.
        opts = append(opts, 
            option.WithEndpoint(pubsubEmulatorHost),
            option.WithoutAuthentication(),
        )
    } else {
        log.Println("PUBSUB_EMULATOR_HOST not set; using credentials")
        opts = append(opts, option.WithCredentialsFile("pubsub-credentials.json"))
    }

    pubsubClient, err := pubsub.NewClient(ctx, projectID, opts...)
    if err != nil {
        return nil, nil, err
    }

    // create topic (algolia and bigquery both subscribe to this topic)
    applicationsTopic, err := pubsubClient.CreateTopic(ctx, "applications")
    if err != nil {
        if err.Error() == "rpc error: code = AlreadyExists desc = Topic already exists" {
            applicationsTopic = pubsubClient.Topic("applications")
            fmt.Println("applications topic already exists, connecting to it")
        } else {
            return nil, nil,  err
        }
    }

	applicationsTopic.EnableMessageOrdering = true

    return pubsubClient, applicationsTopic, nil
}