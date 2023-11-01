package service

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"sys.io/challenge-service/config"
	"sys.io/challenge-service/utils"
)

// type Callback func(ch *amqp.Channel, ctx context.Context, msg []byte, routingKey string, eventName string)

func Consume(rmq *config.RabbitMQ, queueName string) {
	// Create a new channel for this queue
	ch, err := rmq.Conn.Channel()
	utils.FailOnError(err, "Failed to open a channel")

	defer ch.Close()

	msgs, err := ch.Consume(
		queueName,
		"challengeService", // consumer
		false,              // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	utils.FailOnError(err, fmt.Sprintf("Failed to register a consumer for queue %s: %s", queueName, err))

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			ctx := context.Background()

			// Process the message
			log.Printf("Received a message from queue %s: %s", queueName, d.Body)

			// Process message based on Routing Key

			routingKey := utils.GetSuffix(d.RoutingKey)

			if routingKey == "challengeCreate" {
				newRoutingKey := "challengeCreated"
				CreateChallenge(ch, ctx, d.Body, newRoutingKey)
			} else if routingKey == "challengeStart" {
				newRoutingKey := "challengeStarted"
				StartChallenge(ch, ctx, d.Body, newRoutingKey)
			}

			// Acknowledge the message
			err = d.Ack(false)
			utils.FailOnError(err, "Failed to ack")
		}
	}()

	log.Printf(" [*] Waiting for messages")
	<-forever
}

func Publish(ch *amqp.Channel, ctx context.Context, msg []byte, routingKey string) {

	err := ch.PublishWithContext(
		ctx,
		"topic.challenge", // exchange
		fmt.Sprintf("challenge.fromService.%s", routingKey), // routing key
		true,  // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		})
	utils.FailOnError(err, "Failed to publish a message")
	log.Printf("Published a message with routing key %s", fmt.Sprintf("challenge.fromService.%s", routingKey))
}
