package main

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"sys.io/challenge-service/app"
	"sys.io/challenge-service/config"
	"sys.io/challenge-service/utils"
)

func main() {

	// init env
	log.Println("Loadingg .env file")
	config.InitEnv()
	log.Println(".env loaded!")

	rmq := config.SetupMQ()
	defer rmq.Conn.Close()
	defer rmq.Ch.Close()

	msgs, err := rmq.Ch.Consume(
		"queue.challenge.toService", // queue
		"challengeService",          // consumer
		false,                       // auto-ack
		false,                       // exclusive
		false,                       // no-local
		false,                       // no-wait
		nil,                         // args
	)

	utils.FailOnError(err, "Failed to consume messages from queue.challenge.toService")

	var forever chan struct{}

	go func() {

		for msg := range msgs {
			log.Println("Consuming")
			log.Printf("Received msg: %s\n", msg.Body)

			// Decode the JSON message body.
			var data map[string]interface{}
			err := json.Unmarshal(msg.Body, &data)
			if err != nil {
				log.Printf("Failed to decode JSON message body: %s", err)
				continue
			}

			// Retrieve the repository, tag, and release_id from the JSON data.
			repository, tag, release_id := data["repository"].(string), data["tag"].(string), data["release_id"].(string)

			privKey, publicIPAdress, nodePort, err := app.CreateChallenge(repository, tag, release_id)

			if err != nil {
				log.Printf("Failed to create challenge: %s", err)
				continue
			}

			msgBody, err := json.Marshal(map[string]interface{}{
				"message":        "Challenge created successfully.",
				"privKey":        privKey,
				"publicIPAdress": publicIPAdress,
				"nodePort":       nodePort,
			})
			if err != nil {
				log.Printf("Failed to marshal JSON message body: %s", err)
				continue
			}

			q, err := rmq.Ch.QueueDeclare(
				"queue.challenge.FromService", // name
				true,                          // durable
				false,                         // delete when unused
				false,                         // exclusive
				false,                         // no-wait
				nil,                           // arguments
			)
			if err != nil {
				log.Printf("Failed to marshal JSON message body: %s", err)
				continue
			}
			// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			// defer cancel()

			err = rmq.Ch.PublishWithContext(context.TODO(),
				"",     // exchange
				q.Name, // routing key
				false,  // mandatory
				false,  // immediate
				amqp.Publishing{
					ContentType: "application/json",
					Body:        msgBody,
				},
			)
			if err != nil {
				log.Printf("Failed to publish message to queue.challenge.FromService: %s", err)
				continue
			}

			err = msg.Ack(false)
			if err != nil {
				log.Printf("Failed to ack message: %s", err)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages")
	<-forever
}
