package database

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
	"yacoid_server/common"
	"yacoid_server/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateAuthor(request *types.CreateAuthorRequest, authToken string) error {

	user, userError := GetUserByAuthToken(authToken)

	if userError != nil {
		return userError
	}

	var author types.Author

	author.ID = primitive.NewObjectID()
	author.SlugId = fmt.Sprintf("%s-%s-%08d", strings.ToLower(request.LastName), strings.ToLower(request.FirstName), rand.Intn(10000000))
	author.SubmittedBy = user.ID
	author.SubmittedDate = time.Now()

	_, err := authorsCollection.InsertOne(dbContext, author)

	if err != nil {
		return err
	}

	return nil

}

func GetAuthorById(stringId string) (*types.Author, error) {

	id, idError := primitive.ObjectIDFromHex(stringId)

	if idError != nil {
		return nil, idError
	}

	return GetAuthor(id)

}

func GetAuthor(id primitive.ObjectID) (*types.Author, error) {

	filter := bson.M{"_id": id}

	result := authorsCollection.FindOne(dbContext, filter)

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, common.ErrorNotFound
		}
		return nil, result.Err()
	}

	var author types.Author
	decodeError := result.Decode(&author)

	if decodeError != nil {
		return nil, decodeError
	}

	return &author, nil

}
