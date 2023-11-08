package service

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"sys.io/challenge-service/config"
	"sys.io/challenge-service/utils"
)

// type Callback func(ch *amqp.Channel, ctx context.Context, msg []byte, routingKey string, eventName string)

func connectToRabbitMQ(rmq *config.RabbitMQ) (*amqp.Connection, error) {
	// Connect to MQ
	log.Println("Connecting to MQ")
	conn, err := amqp.Dial(rmq.Url)
	if err != nil {
		return nil, err
	}
	log.Println("Connected to MQ!")

	return conn, nil
}

func establishConnection(rmq *config.RabbitMQ, queueName string) (*amqp.Channel, <-chan amqp.Delivery, error) {

	// Create a new connection
	conn, err := connectToRabbitMQ(rmq)
	if err != nil {
		return nil, nil, err
	}

	// Create a new channel for this queue
	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	msgs, err := ch.Consume(
		queueName,
		"processEngine", // consumer
		false,           // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		ch.Close()
		return nil, nil, fmt.Errorf("failed to register a consumer for queue %s: %v", queueName, err)
	}

	log.Println("Consuming from queue: " + queueName)

	return ch, msgs, nil
}

func Consume(rmq *config.RabbitMQ, queueName string) {

	forever := make(chan bool)

	go func() {
		for {
			ch, msgs, err := establishConnection(rmq, queueName)
			if err != nil {
				log.Printf("Failed to establish connection: %s", err)
				time.Sleep(time.Second * 5) // Wait before trying to reconnect
				continue
			}

			notify := ch.NotifyClose(make(chan *amqp.Error))
		
		consumeLoop:
			for {
				select {
				case err := <-notify:
					if err != nil {
						log.Printf("Channel closed for queue %s: %s", queueName, err)
					}
					ch.Close()
					break consumeLoop
				case d := <-msgs:
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
			}
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
