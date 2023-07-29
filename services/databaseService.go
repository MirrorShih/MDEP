package services

import (
	"context"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

type DetectorRes struct {
	Id   primitive.ObjectID `json:"detector_id" bson:"_id,omitempty"`
	Name string             `json:"detector_name" bson:"name,omitempty"`
}

type ReportRes struct {
	Id             primitive.ObjectID `json:"report_id" bson:"_id,omitempty"`
	FuncType       string             `json:"function_type" bson:"function_type,omitempty"`
	Accuracy       float64            `json:"accuracy" bson:"accuracy,omitempty"`
	FP             float64            `json:"fp" bson:"fp,omitempty"`
	FN             float64            `json:"fn" bson:"fn,omitempty"`
	Precision      float64            `json:"precision" bson:"precision,omitempty"`
	Recall         float64            `json:"recall" bson:"recall,omitempty"`
	F1             float64            `json:"f1" bson:"f1,omitempty"`
	TestTime       float64            `json:"testing_time" bson:"testing_time,omitempty"`
	TestSampleNum  float64            `json:"testing_sample_num" bson:"testing_sample_num,omitempty"`
	TotalSampleNum float64            `json:"total_sample_num" bson:"total_sample_num,omitempty"`
}

var MongoClient = NewMongoDBClient(os.Getenv("MONGODB_URI"))

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

func (mongoClient *MongoDBClient) InsertDetector(databaseName string, collectionName string, d DetectorRes) bool {
	collection := mongoClient.client.Database(databaseName).Collection(collectionName)
	_, err := collection.InsertOne(MongoClient.ctx, d)
	if err != nil {
		return false
	}
	log.Println("detector inserted")
	return true
}

func (mongoClient *MongoDBClient) ListDetector(databaseName string, collectionName string) []DetectorRes {
	collection := mongoClient.client.Database(databaseName).Collection(collectionName)
	cursor, err := collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		log.Println(err.Error())
	}
	var results []DetectorRes
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	if err != nil {
		log.Println(err.Error())
	}
	return results
}

func (mongoClient *MongoDBClient) PatchDetector(databaseName string, collectionName string, filter bson.D) bool {
	collection := mongoClient.client.Database(databaseName).Collection(collectionName)
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(mongoClient.ctx, filter, opts)
	if err != nil {
		return false
	}
	return true
}

func (mongoClient *MongoDBClient) GetCertainDetector(databaseName string, collectionName string, filter bson.D) *DetectorRes {
	collection := mongoClient.client.Database(databaseName).Collection(collectionName)
	var result *DetectorRes
	err := collection.FindOne(mongoClient.ctx, filter).Decode(&result)
	if err != nil {
		log.Println("find data failed")
		log.Println(err.Error())
	}
	return result
}

func (mongoClient *MongoDBClient) DeleteDetector(databaseName string, collectionName string, filter bson.D) bool {
	collection := mongoClient.client.Database(databaseName).Collection(collectionName)
	_, err := collection.DeleteOne(MongoClient.ctx, filter)
	if err != nil {
		return false
	}
	log.Println("report inserted")
	return true
}

func (mongoClient *MongoDBClient) InsertReport(databaseName string, collectionName string, report ReportRes) bool {
	collection := mongoClient.client.Database(databaseName).Collection(collectionName)
	_, err := collection.InsertOne(MongoClient.ctx, report)
	if err != nil {
		return false
	}
	log.Println("report inserted")
	return true
}

func (mongoClient *MongoDBClient) ListReport(databaseName string, collectionName string) []ReportRes {
	collection := mongoClient.client.Database(databaseName).Collection(collectionName)
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Println(err.Error())
	}
	var results []ReportRes
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	if err != nil {
		log.Println(err.Error())
	}
	return results
}

func (mongoClient *MongoDBClient) GetCertainReport(databaseName string, collectionName string, filter bson.D) *ReportRes {
	collection := mongoClient.client.Database(databaseName).Collection(collectionName)
	var result *ReportRes
	err := collection.FindOne(mongoClient.ctx, filter).Decode(&result)
	if err != nil {
		log.Println("find data failed")
		log.Println(err.Error())
	}
	return result
}
