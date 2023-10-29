package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	KUBECONFIG         string
	ENVIRONMENT        string
	HELM_CHART_NAME    string
	HELM_REPO_NAME     string
	HELM_REPO_URL      string
	HELM_REPO_USERNAME string
	HELM_REPO_PASSWORD string
	// RABBITMQ_USERNAME  string
	// RABBITMQ_PASSWORD  string
	MONGODB_URL string

	AMQP_URL string
)

func InitEnv() {
	// loads environment variables
	err := godotenv.Load("/app/secrets/.env")
	if err != nil {
		log.Printf("Failed to load secrets/.env")
	}

	// env type
	ENVIRONMENT = os.Getenv("ENVIRONMENT")

	// kube env
	KUBECONFIG = os.Getenv("KUBECONFIG")
	if KUBECONFIG == "" {
		KUBECONFIG = fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))
	}
	log.Printf("KUBECONFIG=%s", KUBECONFIG)

	// helm env
	HELM_CHART_NAME = os.Getenv("HELM_CHART_NAME")
	HELM_REPO_NAME = os.Getenv("HELM_REPO_NAME")
	HELM_REPO_URL = os.Getenv("HELM_REPO_URL")
	HELM_REPO_USERNAME = os.Getenv("HELM_REPO_USERNAME")
	HELM_REPO_PASSWORD = os.Getenv("HELM_REPO_PASSWORD")

	// rmq env
	// RABBITMQ_USERNAME = os.Getenv("RABBITMQ_USERNAME")
	// RABBITMQ_PASSWORD = os.Getenv("RABBITMQ_PASSWORD")
	amqpUsername := os.Getenv("AMQP_USERNAME")
	amqpPassword := os.Getenv("AMQP_PASSWORD")
	amqpHostname := os.Getenv("AMQP_HOSTNAME")
	AMQP_URL = fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		amqpUsername,
		amqpPassword,
		amqpHostname,
		"5672",
	)
	fmt.Printf("AMQP LINK: %s", AMQP_URL)

	// mongo env
	user := os.Getenv("MONGODB_USERNAME")
	pass := os.Getenv("MONGODB_PASSWORD")
	host := os.Getenv("MONGODB_HOSTNAME")
	port := "27017"
	MONGODB_URL = fmt.Sprintf("mongodb://%s:%s@%s:%s", user, pass, host, port)
}

func GetMongoURI() string {
	err := godotenv.Load("secrets/.env")
	if err != nil {
		panic("Error loading env file")
	}
	MONGO_USER := os.Getenv("MONGODB_USERNAME")
	MONGO_PASS := os.Getenv("MONGODB_PASSWORD")
	MONGO_HOSTNAME := os.Getenv("MONGODB_HOSTNAME")
	if MONGODB_URL == "" {
		MONGODB_URL = fmt.Sprintf("mongodb://%s:%s@%s:%s", MONGO_USER, MONGO_PASS, MONGO_HOSTNAME, "27017")
	}
	return MONGODB_URL
}

