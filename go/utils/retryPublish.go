package utils

import (
	"log"
	"math"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishWithRetry(ch *amqp.Channel, exchange, routingKey string, mandatory, immediate bool, msg amqp.Publishing) error {
    const maxRetries = 5

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
        // Log and wait before the next retry (exponential backoff)
        delay := time.Duration(math.Pow(2, float64(i))) * time.Second
        log.Printf("Error publishing message: %v. Retrying in %v...", err, delay)
        time.Sleep(delay)
    }
    return lastErr
}