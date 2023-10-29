package collections

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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

func UpdateAttempt(attempt *models.Attempt) (updatedAttempt *models.Attempt, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	// Create a filter to find the document that you want to update.
	token := attempt.Token
	filter := bson.D{{Key: "token", Value: token}}


	// Create an 4update document to update the value of the object.
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "ipaddress", Value: attempt.Ipaddress},{Key: "port", Value: attempt.Port},{Key: "sshkey", Value: attempt.Sshkey}}}}
	err = attemptCollection.FindOneAndUpdate(ctx, filter, update).Decode(&attempt)

	return attempt, err
}