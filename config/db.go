package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client = DBinstance()

func DBinstance() (client *mongo.Client) {
	MONGODB_URL = GetMongoURI()

	log.Printf("Attempting connection with: %s\n", MONGODB_URL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MONGODB_URL))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	log.Println("Success!")

	log.Println("Pinging server ...")
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping cluster: %v", err)
	}
	log.Println("Success!")

	// initialise indexes
	log.Println("Initialising indexes ...")
	InitIndexes(client)
	log.Println("Success!")
	return client
}

func InitIndexes(client *mongo.Client) {

	// cob_challenge_1 index
	challengeCollection := OpenCollection(client, "challenge")

	challengeIndexModel := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "corId", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "challengeName", Value: 1},
				{Key: "creatorName", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}
	challengeIndexCreated, err := challengeCollection.Indexes().CreateMany(context.Background(), challengeIndexModel)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Created Challenge Index %s\n", challengeIndexCreated)

	// cob_attempt_1 index
	attemptCollection := OpenCollection(client, "attempt")

	attemptIndexModel := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "token", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "challengeName", Value: 1},
				{Key: "creatorName", Value: 1},
				{Key: "participant", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}
	attemptIndexCreated, err := attemptCollection.Indexes().CreateMany(context.Background(), attemptIndexModel)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Created Attempt Index %s\n", attemptIndexCreated)

}

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {

	var collection *mongo.Collection = client.Database("cob").Collection(collectionName)

	return collection
}
