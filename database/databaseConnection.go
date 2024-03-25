package database

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBinstance() *mongo.Client {
	// MongoDb := "mongodb://localhost:27017"
	MongoDbUri := os.Getenv("MONGO_DB_URI")
	// fmt.Print(MongoDb)

	// client, err := mongo.NewClient(options.Client().ApplyURI(MongoDb))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// defer cancel()

	// err = client.Connect(ctx)

	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("connected to mongodb")
	// return client
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(MongoDbUri).SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	// Send a ping to confirm a successful connection
	var result bson.M
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	return client
}

var Client *mongo.Client = DBinstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection *mongo.Collection = client.Database("restaurant").Collection(collectionName)

	return collection
}
