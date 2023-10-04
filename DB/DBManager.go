package DB

import (
	"context"
	"log"

	// get an object type
	// "encoding/json"
	"server/Env"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	dbURL  = Env.Reader.DB_URL
	dbName = Env.Reader.DB_NAME
)

var Collections DBCollections

type DBCollections struct {
	User    *mongo.Collection
	Project *mongo.Collection
	Beacon  *mongo.Collection
}

func InitCollections() bool {
	var err error
	Collections.User, err = GetMongoDbCollection(dbName, "user")
	if err != nil {
		return false
	}
	Collections.Project, err = GetMongoDbCollection(dbName, "project")
	if err != nil {
		return false
	}
	Collections.Beacon, err = GetMongoDbCollection(dbName, "beacon")
	if err != nil {
		return false
	}
	return err == nil
}

// GetMongoDbConnection get connection of mongodb
func getMongoDbConnection() (*mongo.Client, error) {

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(dbURL))

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	return client, nil
}

func GetMongoDbCollection(DbName string, CollectionName string) (*mongo.Collection, error) {
	client, err := getMongoDbConnection()

	if err != nil {
		return nil, err
	}

	collection := client.Database(DbName).Collection(CollectionName)

	return collection, nil
}
