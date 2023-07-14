package services

import (
	"context"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
)

type MongoDBClient struct {
	client *mongo.Client
	ctx    context.Context
}

var mongoClient = NewMongoDBClient(os.Getenv("MONGODB_URI"))

func NewMongoDBClient(uri string) *MongoDBClient {
	client, ctx, err := ConnectMongo(uri)
	if err != nil {
		log.Println("connect mongodb failed")
		log.Println(err.Error())
	}
	if PingMongo(client, ctx) != nil {
		log.Println("ping failed")
		log.Println(err.Error())
	}
	return &MongoDBClient{client, ctx}
}

func ConnectMongo(uri string) (*mongo.Client, context.Context, error) {
	ctx := context.Background()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	return client, ctx, err
}

func CloseMongo(client *mongo.Client, ctx context.Context, cancel context.CancelFunc) {
	defer cancel()

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

func PingMongo(client *mongo.Client, ctx context.Context) error {

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}
	log.Println("connected successfully")
	return nil
}
