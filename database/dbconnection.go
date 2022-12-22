package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Instancedb() *mongo.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error in .env file")
	}
	uri := os.Getenv("MONGODB_URL")
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errs := client.Connect(ctx)
	if errs != nil {
		log.Fatal(errs)
	}
	return client
}

var Client *mongo.Client = Instancedb()

func OpenCollection(client *mongo.Client, collectionname string) *mongo.Collection {
	var collection *mongo.Collection = client.Database("cluster0").Collection(collectionname)
	return collection
}
