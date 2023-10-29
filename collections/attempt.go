package collections

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"sys.io/challenge-service/config"
	"sys.io/challenge-service/models"
)

var attemptCollection *mongo.Collection = config.OpenCollection(config.Client, "attempt")

func CreateAttempt(attempt *models.Attempt) (result *mongo.InsertOneResult, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	return attemptCollection.InsertOne(ctx, attempt)
}