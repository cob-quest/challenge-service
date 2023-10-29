package collections

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"sys.io/challenge-service/config"
	"sys.io/challenge-service/models"
)

var imageCollection *mongo.Collection = config.OpenCollection(config.Client, "image_builder")

func GetImage(creatorName ,imageName, imageTag string) (result models.Image, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()


	var image models.Image

	filter := bson.D{{Key: "creatorName", Value: creatorName}, {Key: "imageName", Value: imageName}, {Key: "imageTag", Value: imageTag}}
	err = imageCollection.FindOne(ctx, filter).Decode(&image)

	return image, err

}