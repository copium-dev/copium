package main

// simple consumer; for now just receive and print
// later, use worker pools (goroutines) to handle messages to index algolia
import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

// idk im tired of having to write log.Fatalf so why not just make a function
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	// connects to rabbitmq server (make sure you ran below command first lol)
	// docker run -d --hostname my-rabbit --name rabbit -p 5672:5672 -p 15672:15672 rabbitmq:3-management
	// to close:
	//	docker stop rabbit
	// 	docker rm rabbit
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
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
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

	// consume messages from the queue
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	// log readiness, wait indefinitely (until manually killed or interrupted)
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<- forever
}