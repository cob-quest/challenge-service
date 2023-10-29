package collections

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"sys.io/challenge-service/config"
	"sys.io/challenge-service/models"
)

var challengeCollection *mongo.Collection = config.OpenCollection(config.Client, "challenge")

func CreateChallenge(challenge *models.Challenge) (result *mongo.InsertOneResult, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	return challengeCollection.InsertOne(ctx, challenge)
}