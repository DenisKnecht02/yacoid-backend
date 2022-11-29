package database

import (
	"time"
	"yacoid_server/common"
	"yacoid_server/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateSource(request *types.CreateSourceRequest, userId string) error {

	var source types.Source

	source.ID = primitive.NewObjectID()
	source.SubmittedBy = userId
	source.SubmittedDate = time.Now()
	source.Title = request.Title
	source.Type = request.Type

	authors, idError := stringsToObjectIDs(&request.Authors)

	if idError != nil {
		return idError
	}

	authorsExistError := validateAuthorsExist(&authors)

	if authorsExistError != nil {
		return authorsExistError
	}

	source.Authors = authors
	source.BookProperties = request.BookProperties
	source.JournalProperties = request.JournalProperties
	source.WebProperties = request.WebProperties

	_, err := sourcesCollection.InsertOne(dbContext, source)

	if err != nil {
		return err
	}

	return nil

}

func stringsToObjectIDs(stringIds *[]string) ([]primitive.ObjectID, error) {

	ids := []primitive.ObjectID{}
	for _, stringId := range *stringIds {

		id, idError := primitive.ObjectIDFromHex(stringId)

		if idError != nil {
			return nil, idError
		}

		ids = append(ids, id)
	}

	return ids, nil

}

func validateAuthorsExist(ids *[]primitive.ObjectID) error {

	for _, id := range *ids {

		_, err := GetAuthor(id)

		if err != nil {
			return err
		}

	}

	return nil

}

func GetSourceById(stringId string) (*types.Source, error) {

	id, idError := primitive.ObjectIDFromHex(stringId)

	if idError != nil {
		return nil, idError
	}

	return GetSource(id)

}

func GetSource(id primitive.ObjectID) (*types.Source, error) {

	filter := bson.M{"_id": id}

	result := sourcesCollection.FindOne(dbContext, filter)

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, common.ErrorNotFound
		}
		return nil, result.Err()
	}

	var source types.Source
	decodeError := result.Decode(&source)

	if decodeError != nil {
		return nil, decodeError
	}

	return &source, nil

}

func GetSourceCount() (int64, error) {

	count, err := sourcesCollection.CountDocuments(dbContext, bson.M{}, nil)
	return count, err

}

func GetSourceCountInCurrentQuarter() (int64, error) {

	currentQuarterDate := common.GetCurrentQuarterDate()

	filter := bson.M{
		"submitted_date": bson.M{
			"$gte": currentQuarterDate,
		},
	}

	count, err := sourcesCollection.CountDocuments(dbContext, filter, nil)
	return count, err

}
