package services

import (
	"MDEP/models"
	"context"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"io/ioutil"
	"log"
	"os"
)

type MongoDBClient struct {
	client *mongo.Client
	ctx    context.Context
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

func (mongoClient *MongoDBClient) InsertDetector(databaseName, file, filename string, collectionName string) bool {
	detectorID := mongoClient.UploadFile(databaseName, file, filename)
	collection := mongoClient.client.Database(databaseName).Collection(collectionName)
	_, err := collection.InsertOne(MongoClient.ctx, models.DetectorRes{primitive.NewObjectID(), filename, detectorID})
	if err != nil {
		return false
	}
	log.Println("detector inserted")
	return true
}

func (mongoClient *MongoDBClient) ListDetector(databaseName string, collectionName string) []models.DetectorRes {
	collection := mongoClient.client.Database(databaseName).Collection(collectionName)
	cursor, err := collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		log.Println(err.Error())
	}
	var results []models.DetectorRes
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	if err != nil {
		log.Println(err.Error())
	}
	return results
}

func (mongoClient *MongoDBClient) PatchDetector(databaseName string, collectionName string, filter bson.D) bool {
	// TODO
	collection := mongoClient.client.Database(databaseName).Collection(collectionName)
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(mongoClient.ctx, filter, opts)
	if err != nil {
		return false
	}
	return true
}

func (mongoClient *MongoDBClient) GetCertainDetector(databaseName string, collectionName string, filter bson.D) *models.DetectorRes {
	collection := mongoClient.client.Database(databaseName).Collection(collectionName)
	var result *models.DetectorRes
	err := collection.FindOne(mongoClient.ctx, filter).Decode(&result)
	if err != nil {
		log.Println("find data failed")
		log.Println(err.Error())
	}
	return result
}

func (mongoClient *MongoDBClient) DeleteDetector(databaseName string, collectionName string, filter bson.D) bool {
	collection := mongoClient.client.Database(databaseName).Collection(collectionName)
	fileId := mongoClient.GetCertainDetector(databaseName, collectionName, filter).FileId
	mongoClient.DeleteFile(databaseName, fileId)
	_, err := collection.DeleteOne(MongoClient.ctx, filter)
	if err != nil {
		return false
	}
	log.Println("report inserted")
	return true
}

func (mongoClient *MongoDBClient) InsertReport(databaseName string, collectionName string, report models.ReportRes) bool {
	collection := mongoClient.client.Database(databaseName).Collection(collectionName)
	_, err := collection.InsertOne(MongoClient.ctx, report)
	if err != nil {
		return false
	}
	log.Println("report inserted")
	return true
}

func (mongoClient *MongoDBClient) ListReport(databaseName string, collectionName string) []models.ReportRes {
	collection := mongoClient.client.Database(databaseName).Collection(collectionName)
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Println(err.Error())
	}
	var results []models.ReportRes
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	if err != nil {
		log.Println(err.Error())
	}
	return results
}

func (mongoClient *MongoDBClient) GetCertainReport(databaseName string, collectionName string, filter bson.D) *models.ReportRes {
	collection := mongoClient.client.Database(databaseName).Collection(collectionName)
	var result *models.ReportRes
	err := collection.FindOne(mongoClient.ctx, filter).Decode(&result)
	if err != nil {
		log.Println("find data failed")
		log.Println(err.Error())
	}
	return result
}

func (mongoClient *MongoDBClient) UploadFile(databaseName, file, filename string) primitive.ObjectID {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	bucket, err := gridfs.NewBucket(
		mongoClient.client.Database(databaseName),
	)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	uploadStream, err := bucket.OpenUploadStream(
		filename,
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer uploadStream.Close()

	fileSize, err := uploadStream.Write(data)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	log.Printf("Write file to DB was successful. File size: %d M\n", fileSize)
	return uploadStream.FileID.(primitive.ObjectID)
}

func (mongoClient *MongoDBClient) DeleteFile(databaseName string, fileId primitive.ObjectID) {
	bucket, err := gridfs.NewBucket(
		mongoClient.client.Database(databaseName),
	)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	if err := bucket.Delete(fileId); err != nil {
		panic(err)
	}
}
