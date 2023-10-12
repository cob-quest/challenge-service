package config

import (
	"fmt"
	"log"
	"os"
	"github.com/joho/godotenv"
)

var (
	KUBECONFIG string
	ENVIRONMENT string
	HELM_CHART_NAME    string
	HELM_REPO_NAME     string
	HELM_REPO_URL      string
	HELM_REPO_USERNAME string
	HELM_REPO_PASSWORD string
	RABBITMQ_USERNAME string
	RABBITMQ_PASSWORD string
)

func InitEnv() {
	// loads environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env")
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
	RABBITMQ_USERNAME = os.Getenv("RABBITMQ_USERNAME")
	RABBITMQ_PASSWORD = os.Getenv("RABBITMQ_PASSWORD")


}
