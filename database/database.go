package database

import (
	"errors"
	"fmt"
	"os"
	"yacoid_server/constants"

	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var dbContext context.Context
var client *mongo.Client
var database *mongo.Database

var definitionsCollection *mongo.Collection
var authorsCollection *mongo.Collection
var sourcesCollection *mongo.Collection

var InvalidID = errors.New("INVALID_ID")

var ErrorUserNotFound = errors.New("USER_NOT_FOUND")
var ErrorDefinitionNotFound = errors.New("DEFINITION_NOT_FOUND")
var ErrorNotEnoughPermissions = errors.New("NOT_ENOUGH_PERMISSIONS")

func Connect() error {

	fmt.Println("Connecting to database...")

	dbContext = context.TODO()
	databaseURL := os.Getenv(constants.EnvKeyDatabaseUrl)

	options := options.Client().ApplyURI(databaseURL)

	var err error
	client, err = mongo.Connect(dbContext, options)
	if err != nil {
		fmt.Println("Could not connect to database:")
		return err
	}

	fmt.Println("Pinging database...")
	err = client.Ping(dbContext, nil)

	if err != nil {
		fmt.Println("Could not ping database:")
		return err
	}

	fmt.Println("Successfully connected to database!")
	database = client.Database("YACOID")

	database.CreateCollection(dbContext, "definitions")

	definitionsCollection = database.Collection("definitions")
	definitionsCollection.Indexes().CreateOne(dbContext, mongo.IndexModel{
		Keys: bson.D{{Key: "title", Value: "text"}, {Key: "content", Value: "text"}},
	})

	authorsCollection = database.Collection("authors")
	sourcesCollection = database.Collection("sources")

	return nil
}

type UpdateEntry struct {
	field string
	value any
}

type UpdateState struct {
	Success bool    `bson:"success" json:"success"`
	Error   *string `bson:"error,omitempty" json:"error,omitempty"`
}

func CreateUpdateDocument(inputs []UpdateEntry) bson.D {

	var update bson.D

	for _, input := range inputs {
		if input.value != nil {
			update = append(update, bson.E{Key: input.field, Value: input.value})
		}
	}

	return update

}
