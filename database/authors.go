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
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateAuthor(request *types.CreateAuthorRequest, userId string) error {

	var author types.Author

	author.ID = primitive.NewObjectID()
	author.SlugId = fmt.Sprintf("%s-%s-%08d", strings.ToLower(request.LastName), strings.ToLower(request.FirstName), rand.Intn(10000000))
	author.SubmittedBy = userId
	author.SubmittedDate = time.Now()
	author.Type = request.Type

	_, err := authorsCollection.InsertOne(dbContext, author)

	if err != nil {
		return err
	}

	return nil

}

func GetAuthors(pageSize int, page int, definitionFilter *types.AuthorFilter, sort *interface{}) ([]*types.Author, error) {

	if pageSize <= 0 || page <= 0 {
		return nil, common.ErrorInvalidType
	}

	options := options.FindOptions{}

	if sort != nil {
		options.SetSort(*sort)
	}
	options.SetLimit(int64(pageSize))
	options.SetSkip(int64((page - 1) * pageSize))

	filter := CreateAuthorFilterQuery(definitionFilter)
	fmt.Println("FILTER_QUERY")
	fmt.Println(filter)
	return getDocuments[types.Author](authorsCollection, filter, &options)

}

func CreateAuthorFilterQuery(filter *types.AuthorFilter) bson.D {

	query := bson.D{}

	if filter == nil {
		return query
	}

	textSearch := ""
	if filter.FirstName != nil && len(*filter.FirstName) > 0 {
		textSearch = *filter.FirstName
	}

	if filter.LastName != nil && len(*filter.LastName) > 0 {
		if len(textSearch) > 0 {
			textSearch += " "
		}
		textSearch += *filter.LastName
	}

	if filter.OrganizationName != nil && len(*filter.OrganizationName) > 0 {
		if len(textSearch) > 0 {
			textSearch += " "
		}
		textSearch += *filter.OrganizationName
	}

	if len(textSearch) > 0 {
		query = append(query, bson.E{Key: "$text", Value: bson.D{{Key: "$search", Value: textSearch}}})
	}

	if filter.Type != nil {
		query = append(query, bson.E{Key: "type", Value: bson.D{{Key: "$in", Value: *filter.Type}}})
	}

	return query

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
			return nil, common.ErrorAuthorNotFound
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

func GetAuthorCount() (int64, error) {

	count, err := authorsCollection.CountDocuments(dbContext, bson.M{}, nil)
	return count, err

}

func GetAuthorCountInCurrentQuarter() (int64, error) {

	currentQuarterDate := common.GetCurrentQuarterDate()

	filter := bson.M{
		"submitted_date": bson.M{
			"$gte": currentQuarterDate,
		},
	}

	count, err := authorsCollection.CountDocuments(dbContext, filter, nil)
	return count, err

}

func GetAuthorPageCount(request *types.AuthorPageCountRequest) (int64, error) {
	filter := CreateAuthorFilterQuery(request.Filter)
	return getPageCount(authorsCollection, request.PageSize, filter)
}
