package utils

import (
	"log"
	"math"
	"time"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishWithRetry(ch *amqp.Channel, exchange, routingKey string, mandatory, immediate bool, msg amqp.Publishing) error {
    const maxRetries = 3

    var lastErr error
    for i := 0; i < maxRetries; i++ {
        err := ch.Publish(
            exchange,
            routingKey,
            mandatory,
            immediate,
            msg,
        )
        if err == nil {
            // success
            return nil
        }

        lastErr = err
        // log and wait before the next retry (exponential backoff)
        delay := time.Duration(math.Pow(2, float64(i))) * time.Second
        log.Printf("Error publishing message: %v. Retrying in %v...", err, delay)
        time.Sleep(delay)
		
    }
    return lastErr
}

// yes: a pointer to a pointer. this is required because 
// this is trying to modify the original pointer, not the value
// also, amqp.Queue is a struct so we pass a pointer to this struct
func RetryRabbitConnection(ch **amqp.Channel, q *amqp.Queue) error {
    const maxAttempts = 3

    var conn *amqp.Connection
    var localCh *amqp.Channel
    var localQ amqp.Queue
    var lastErr error

    for i := 0; i < maxAttempts; i++ {
        conn, lastErr = amqp.Dial("amqp://guest:guest@localhost:5672/")
        if lastErr != nil {
            // failed to connect to RabbitMQ
            delay := time.Duration(math.Pow(2, float64(i))) * time.Second
            log.Printf("Error connecting to RabbitMQ: %v. Retrying in %v...", lastErr, delay)
            time.Sleep(delay)
            continue
        }

        // open a channel
        localCh, lastErr = conn.Channel()
        if lastErr != nil {
            conn.Close()
            delay := time.Duration(math.Pow(2, float64(i))) * time.Second
            log.Printf("Error opening channel: %v. Retrying in %v...", lastErr, delay)
            time.Sleep(delay)
            continue
        }

        // declare queue
        localQ, lastErr = localCh.QueueDeclare(
            "my-rabbit", // name
            false,       // durable
            false,       // delete when unused
            false,       // exclusive
            false,       // no-wait
            nil,         // arguments
        )
        if lastErr != nil {
            localCh.Close()
            conn.Close()
            delay := time.Duration(math.Pow(2, float64(i))) * time.Second
            log.Printf("Error declaring queue: %v. Retrying in %v...", lastErr, delay)
            time.Sleep(delay)
            continue
        }

        // success: set the pointers
        *ch = localCh
        *q = localQ
        return nil
    }

    return fmt.Errorf("unable to connect to RabbitMQ after %d attempts: %w", maxAttempts, lastErr)
}