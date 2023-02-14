package database

import (
	"fmt"
	"math/rand"
	"time"
	"yacoid_server/auth"
	"yacoid_server/common"
	"yacoid_server/constants"
	"yacoid_server/types"

	"github.com/go-playground/validator/v10"
	"github.com/gosimple/slug"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AuthorToResponse(author *types.Author) (*types.AuthorResponse, error) {

	response := types.AuthorResponse{}
	response.ID = author.ID
	response.SlugId = author.SlugId

	response.SubmittedBy = author.SubmittedBy
	nickname, err := auth.GetNicknameOfUser(author.SubmittedBy)

	if err == nil {
		response.SubmittedByName = nickname
	} else {
		response.SubmittedByName = "<deleted>"
	}

	response.SubmittedDate = author.SubmittedDate
	response.Type = author.Type
	response.PersonProperties = author.PersonProperties
	response.OrganizationProperties = author.OrganizationProperties

	return &response, nil

}

func AuthorsToResponses(authors *[]*types.Author) (*[]types.AuthorResponse, error) {

	responses := []types.AuthorResponse{}

	for _, author := range *authors {

		response, err := AuthorToResponse(author)

		if err != nil {
			return nil, err
		}

		responses = append(responses, *response)
	}

	return &responses, nil

}

func CreateAuthor(request *types.CreateAuthorRequest, userId string) (*primitive.ObjectID, error) {

	var author types.Author

	author.ID = primitive.NewObjectID()

	if request.Type == types.EnumAuthorType.Person && request.PersonProperties != nil {
		text := fmt.Sprintf("%s-%s-%08d", request.PersonProperties.LastName, request.PersonProperties.FirstName, rand.Intn(10000000))
		author.SlugId = slug.Make(text)
		author.PersonProperties = request.PersonProperties
	} else if request.Type == types.EnumAuthorType.Organization && request.OrganizationProperties != nil {
		text := fmt.Sprintf("%s-%08d", request.OrganizationProperties.OrganizationName, rand.Intn(10000000))
		author.SlugId = slug.Make(text)
		author.OrganizationProperties = request.OrganizationProperties
	} else {
		return nil, constants.ErrorAuthorCreation
	}

	now := time.Now()

	author.SubmittedBy = userId
	author.SubmittedDate = now
	author.LastChangeDate = now
	author.ApprovedBy = nil
	author.ApprovedDate = nil
	author.Approved = false

	author.Type = request.Type

	_, err := authorsCollection.InsertOne(dbContext, author)

	if err != nil {
		return nil, err
	}

	return &author.ID, nil

}

/* If the author is used in sources, then an array of the sources and an error will be returned. */
func DeleteAuthor(authorId string) (*[]string, error) {

	id, err := primitive.ObjectIDFromHex(authorId)

	if err != nil {
		return nil, constants.ErrorInvalidID
	}

	sources, err := GetSourcesWithAuthor(id)

	if err != nil {
		return nil, err
	}
	if len(*sources) > 0 {

		sourceIds := []string{}

		for _, source := range *sources {
			sourceIds = append(sourceIds, source.ID.Hex())
		}

		return &sourceIds, constants.ErrorAuthorDeletionBecauseInUse

	}

	filter := bson.M{
		"_id": id,
	}

	result, err := authorsCollection.DeleteOne(dbContext, filter)

	if err != nil {
		return nil, err
	}

	if result.DeletedCount == 0 {
		return nil, constants.ErrorAuthorNotFound
	}

	return nil, nil

}

func ChangeAuthor(request *types.ChangeAuthorRequest, userId string, validate *validator.Validate) error {

	id, err := primitive.ObjectIDFromHex(*request.ID)

	if err != nil {
		return constants.ErrorInvalidID
	}

	author, err := GetAuthor(id)

	if err != nil {
		return err
	}

	if author.Approved {
		return constants.ErrorAuthorAlreadyApproved
	}

	oldAuthor := types.Author{}
	copier.Copy(oldAuthor, author)

	/* Change fields */

	if request.Type != nil {
		author.Type = *request.Type
	}

	if author.Type == types.EnumAuthorType.Person && request.PersonProperties != nil {

		if author.PersonProperties == nil {
			author.PersonProperties = &types.PersonProperties{}
		}

		if request.PersonProperties.FirstName != nil {
			fmt.Printf("Changed firstname from \"%s\" to \"%s\"\n", author.PersonProperties.FirstName, *request.PersonProperties.FirstName)
			author.PersonProperties.FirstName = *request.PersonProperties.FirstName
		}

		if request.PersonProperties.LastName != nil {
			fmt.Printf("Changed lastname from \"%s\" to \"%s\"\n", author.PersonProperties.LastName, *request.PersonProperties.LastName)
			author.PersonProperties.LastName = *request.PersonProperties.LastName
		}

		/* Remove old data in other properties */
		if oldAuthor != *author {
			author.OrganizationProperties = nil
		}

	}

	if author.Type == types.EnumAuthorType.Organization && request.OrganizationProperties != nil {

		if author.OrganizationProperties == nil {
			author.OrganizationProperties = &types.OrganizationProperties{}
		}

		if request.OrganizationProperties.OrganizationName != nil {
			fmt.Printf("Changed lastname from \"%s\" to \"%s\"\n", author.OrganizationProperties.OrganizationName, *request.OrganizationProperties.OrganizationName)
			author.OrganizationProperties.OrganizationName = *request.OrganizationProperties.OrganizationName
		}

		/* Remove old data in other properties */
		if oldAuthor != *author {
			author.PersonProperties = nil
		}

	}

	/* Validate */
	errorFields := author.Validate(validate)

	if errorFields != nil {
		return constants.CreateValidationError(&errorFields)
	}

	/* Update existing author */

	author.LastChangeDate = time.Now()

	filter := bson.M{
		"_id": id,
	}

	result, err := authorsCollection.ReplaceOne(dbContext, filter, author, nil)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return constants.ErrorAuthorNotFound
	}

	return nil

}

func ApproveAuthors(authorIds []primitive.ObjectID, userId string) error {

	filter := bson.M{
		"_id": bson.M{
			"$in": authorIds,
		},
		"approved": false,
	}

	update := bson.M{
		"$set": bson.M{
			"approved_by":   userId,
			"approved_date": time.Now(),
			"approved":      true,
		},
	}

	_, err := authorsCollection.UpdateMany(dbContext, filter, update, nil)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return err
	}

	return nil

}

func GetAuthors(request *types.AuthorPageRequest) ([]*types.Author, error) {

	if request.PageSize <= 0 || request.Page <= 0 {
		return nil, constants.ErrorInvalidType
	}

	options := options.FindOptions{}

	options.SetLimit(int64(request.PageSize))
	options.SetSkip(int64((request.Page - 1) * request.PageSize))

	filter := CreateAuthorFilterQuery(request.Filter)
	return getDocuments[types.Author](authorsCollection, filter, &options)

}

func CreateAuthorFilterQuery(filter *types.AuthorFilter) bson.D {

	query := bson.D{}

	if filter == nil {
		return query
	}

	if filter.Name != nil && len(*filter.Name) > 0 {
		query = append(query, bson.E{Key: "$text", Value: bson.D{{Key: "$search", Value: *filter.Name}}})
	}

	if filter.Types != nil && len(*filter.Types) > 0 {
		query = append(query, bson.E{Key: "type", Value: bson.D{{Key: "$in", Value: *filter.Types}}})
	}

	if filter.Approved != nil {
		query = append(query, bson.E{Key: "approved", Value: filter.Approved})
	}

	return query

}

func GetAuthorById(stringId string) (*types.Author, error) {

	id, err := primitive.ObjectIDFromHex(stringId)

	if err != nil {
		return nil, constants.ErrorInvalidID
	}

	return GetAuthor(id)

}

func GetAuthor(id primitive.ObjectID) (*types.Author, error) {

	filter := bson.M{"_id": id}

	result := authorsCollection.FindOne(dbContext, filter)

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, constants.ErrorAuthorNotFound
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

func GetAuthorsByIds(ids *[]primitive.ObjectID) (*[]*types.Author, error) {

	filter := bson.M{"_id": bson.D{{Key: "$in", Value: *ids}}}

	options := options.FindOptions{}
	authors, err := getDocuments[types.Author](authorsCollection, filter, &options)

	if err != nil {
		return nil, err
	}

	return &authors, nil

}

func GetAuthorCount() (int64, error) {

	filter := bson.M{
		"approved": true,
	}

	count, err := authorsCollection.CountDocuments(dbContext, filter, nil)
	return count, err

}

func GetAuthorCountInCurrentQuarter() (int64, error) {

	currentQuarterDate := common.GetCurrentQuarterDate()

	filter := bson.M{
		"approved": true,
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

func validateAuthorsExist(ids *[]primitive.ObjectID) error {

	for _, id := range *ids {

		_, err := GetAuthor(id)

		if err != nil {
			return err
		}

	}

	return nil

}
