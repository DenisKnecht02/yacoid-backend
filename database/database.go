package database

import (
	"fmt"
	"math"
	"os"
	"yacoid_server/constants"

	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var dbContext context.Context
var client *mongo.Client
var database *mongo.Database

var definitionsCollection *mongo.Collection
var authorsCollection *mongo.Collection
var sourcesCollection *mongo.Collection

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
	_, err = definitionsCollection.Indexes().CreateOne(dbContext, mongo.IndexModel{
		Keys: bson.D{{Key: "title", Value: "text"}, {Key: "content", Value: "text"}},
	})

	if err != nil {
		return err
	}

	authorsCollection = database.Collection("authors")
	_, err = authorsCollection.Indexes().CreateOne(dbContext, mongo.IndexModel{
		Keys: bson.D{{Key: "person_properties.first_name", Value: "text"}, {Key: "person_properties.last_name", Value: "text"}, {Key: "organization_properties.organization_name", Value: "text"}},
	})

	if err != nil {
		return err
	}

	sourcesCollection = database.Collection("sources")
	_, err = sourcesCollection.Indexes().CreateOne(dbContext, mongo.IndexModel{
		Keys: bson.D{
			{Key: "book_properties.title", Value: "text"}, {Key: "book_properties.edition", Value: "text"}, {Key: "book_properties.publisher", Value: "text"},
			{Key: "journal_properties.journal_name", Value: "text"}, {Key: "journal_properties.title", Value: "text"}, {Key: "journal_properties.edition", Value: "text"}, {Key: "journal_properties.publisher", Value: "text"},
			{Key: "web_properties.article_name", Value: "text"}, {Key: "web_properties.url", Value: "text"}, {Key: "web_properties.website_name", Value: "text"},
		},
	})

	if err != nil {
		return err
	}

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

func getDocuments[T interface{}](collection *mongo.Collection, filter interface{}, options *options.FindOptions) ([]*T, error) {

	cursor, err := collection.Find(dbContext, filter, options)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(dbContext)

	documents := []*T{}

	for cursor.Next(dbContext) {

		var document T
		err := cursor.Decode(&document)

		if err != nil {
			return nil, err
		}

		documents = append(documents, &document)
	}

	return documents, nil

}

func aggregateDocuments[T interface{}](collection *mongo.Collection, pipeMap interface{}, options *options.AggregateOptions) ([]*T, error) {

	cursor, err := collection.Aggregate(dbContext, pipeMap, options)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(dbContext)

	documents := []*T{}

	for cursor.Next(dbContext) {

		var document T
		err := cursor.Decode(&document)

		if err != nil {
			return nil, err
		}

		documents = append(documents, &document)
	}

	return documents, nil

}

func countDocuments(collection *mongo.Collection, filter interface{}, countOptions *options.CountOptions) (int, error) {

	if countOptions == nil {
		countOptions = options.Count()
	}

	count, err := collection.CountDocuments(dbContext, filter)

	if err != nil {
		return 0, err
	}

	return int(count), nil

}

func getPageCount(collection *mongo.Collection, pageSize int, filter interface{}) (int64, error) {

	count, err := collection.CountDocuments(dbContext, filter, nil)
	pageCount := int64(math.Ceil(float64(count) / float64(pageSize)))

	if err != nil {
		return 0, err
	}

	return pageCount, nil

}

func stringsToObjectIDs(stringIds *[]string) ([]primitive.ObjectID, error) {

	ids := []primitive.ObjectID{}

	if stringIds == nil {
		return ids, nil
	}

	for _, stringId := range *stringIds {

		id, idError := primitive.ObjectIDFromHex(stringId)

		if idError != nil {
			return nil, constants.ErrorInvalidID
		}

		ids = append(ids, id)
	}

	return ids, nil

}

func appendUpdate(key string, currentValue interface{}, newValue interface{}) (entries bson.D) {

	if newValue != currentValue && newValue != nil {
		entries = append(entries, bson.E{Key: key, Value: newValue})
	}

	return entries

}

func forceUpdate(key string, newValue interface{}) (entries bson.D) {

	entries = append(entries, bson.E{Key: key, Value: newValue})
	return entries

}
