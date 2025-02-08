package main

// simple consumer; for now just receive and print
// later, use worker pools (goroutines) to handle messages to index algolia
import (
	"log"
	"os"

	"github.com/juhun32/copium/rabbit-consumer/config"
	"github.com/juhun32/copium/rabbit-consumer/pool"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func initializeAlgoliaClient() (*search.APIClient, error) {
	appID := os.Getenv("ALGOLIA_APP_ID")
	writeApiKey := os.Getenv("ALGOLIA_WRITE_API_KEY")

	algoliaClient, err := search.NewClient(appID, writeApiKey)
	if err != nil {
		return nil, err
	}

	return algoliaClient, nil
}

// docker run -d --hostname my-rabbit --name rabbit -p 5672:5672 -p 15672:15672 rabbitmq:3-management
// to close:
//
//	docker stop rabbit
//	docker rm rabbit
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// create an algolia client (shared across all workers)
	algoliaClient, err := initializeAlgoliaClient()
	if err != nil {
		log.Fatalf("Error initializing algolia client")
	}

	// configure worker pool with num workers, queue name, and algolia client
	cfg := config.NewConfig(10000, "my-rabbit", algoliaClient)

	// connect to rabbit
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// open a channel to communicate with rabbit
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// declare a queue to consume from
	q, err := ch.QueueDeclare(
		"my-rabbit", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// register a consumer
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	// create a channel to keep main function running
	forever := make(chan bool)

	// create worker pool w/ config
	p := pool.NewPool(cfg.NumWorkers, cfg.AlgoliaClient)

	p.Run()

	// run forever; for each message...
	// 1. log
	// 2. send to worker pool's job queue
	//  - a worker will pick up the job and process it. If no available workers, enqueue the job
	go func() {
		var counter int32 = 1
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			// NOTE: We don't want to unmarshal the message here, rather
			// the worker should do this. This is because unmarshaling here
			// will block message receiving so let the worker do it
			p.JobQueue <- pool.Job{
				ID:   counter,
				Data: d.Body,
			}
			counter++
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
