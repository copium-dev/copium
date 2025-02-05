package main

import (
    "log"
    "context"
	"fmt"
    "os"

    "github.com/juhun32/jtracker-backend/cmd/api"
    "github.com/juhun32/jtracker-backend/utils"

	amqp "github.com/rabbitmq/amqp091-go"

    firebase "firebase.google.com/go"
    "google.golang.org/api/option"
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
    if emulatorHost := os.Getenv("FIRESTORE_EMULATOR_HOST"); emulatorHost != "" {
        log.Printf("Connecting to Firestore emulator at %s", emulatorHost)
        conf.DatabaseURL = "http://" + emulatorHost
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

	// initialize rabbit (this is the producer)
	ch, q, err := initializeRabbit()
	if err != nil {
		// actually we have to log.Fatal here since otherwise it would do nil pointer dereference
		log.Fatal("Failed to initialize RabbitMQ: ", err)
	}

    // temp: firestore emulator is on 8080 so use 8000 for API server
    // in prod, use Google Cloud Run's default PORT env variable
    server := api.NewAPIServer(":" + port, firestoreClient, authHandler, ch, q)
    if err := server.Run(); err != nil {
        log.Fatal(err)
    }
}

func initializeRabbit() (*amqp.Channel, amqp.Queue, error) {
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        return nil, amqp.Queue{}, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
    }
    // Don't defer conn.Close() - connection should stay open

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, amqp.Queue{}, fmt.Errorf("failed to open channel: %w", err)
    }

    q, err := ch.QueueDeclare(
        "my-rabbit", // name
        false,       // durable
        false,       // delete when unused
        false,       // exclusive
        false,       // no-wait
        nil,        // arguments
    )
    if err != nil {
        ch.Close()
        conn.Close()
        return nil, amqp.Queue{}, fmt.Errorf("failed to declare queue: %w", err)
    }

    return ch, q, nil
}