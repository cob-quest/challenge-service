package config

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"sys.io/challenge-service/utils"
)

type RabbitMQ struct {
	Conn *amqp.Connection
	Ch   *amqp.Channel
	Url  string
}

func SetupMQ() *RabbitMQ {

	// url := fmt.Sprintf("amqp://%s:%s@rabbitmq-headless.platform.svc.cluster.local:5672", RABBITMQ_USERNAME, RABBITMQ_PASSWORD)

	// if ENVIRONMENT == "DEV" {
	// 	url = fmt.Sprintf("amqp://%s:%s@%s:5672", RABBITMQ_USERNAME, RABBITMQ_PASSWORD,"127.0.0.1")
	// }

	// Connect to MQ
	log.Println("Connecting to MQ")
	conn, err := amqp.Dial(AMQP_URL)
	utils.FailOnError(err, "Failed to connect to RabbitMQ")
	log.Println("Connected to MQ!")

	// Open Channel
	ch, err := conn.Channel()
	utils.FailOnError(err, "Failed to open a channel")

	return &RabbitMQ{
		Conn: conn,
		Ch:   ch,
		Url:  AMQP_URL,
	}
}
