package main

import (
	"log"

	"sys.io/challenge-service/services"
	"sys.io/challenge-service/config"
)

func main() {

	// init env
	log.Println("Loadingg .env file")
	config.InitEnv()
	log.Println(".env loaded!")

	rmq := config.SetupMQ()
	defer rmq.Conn.Close()
	defer rmq.Ch.Close()


	go service.Consume(rmq, "queue.challenge.toService")
	
	select {}
}
