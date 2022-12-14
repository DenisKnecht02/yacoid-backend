package database

import (
	"fmt"
	"time"

	"yacoid_server/common"
	"yacoid_server/constants"
	"yacoid_server/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SubmitDefinition(request *types.SubmitDefinitionRequest, userId string) (*primitive.ObjectID, error) {

	var definition types.Definition

	now := time.Now()
	definition.ID = primitive.NewObjectID()
	definition.SubmittedBy = userId
	definition.SubmittedDate = now
	definition.LastChangeDate = now
	definition.ApprovedBy = nil
	definition.ApprovedDate = nil
	definition.Approved = false

	definition.Title = request.Title
	definition.Content = request.Content
	definition.Category = request.Category
	definition.PublishingDate = request.PublishingDate

	sourceId, err := primitive.ObjectIDFromHex(request.SourceId)

	if err != nil {
		return nil, constants.ErrorInvalidID
	}

	err = validateSourceExists(sourceId)

	if err != nil {
		return nil, err
	}

	definition.Source = sourceId

	rejectionLog := []*types.Rejection{}
	definition.RejectionLog = &rejectionLog

	_, err = definitionsCollection.InsertOne(dbContext, definition)
	// TODO: send email to user??

	if err != nil {
		return nil, err
	}

	return &definition.ID, nil

}

func ApproveDefinition(definitionId string, userId string) error {

	id, err := primitive.ObjectIDFromHex(definitionId)

	if err != nil {
		return constants.ErrorInvalidID
	}

	definition, err := GetDefinitionByObjectId(id)

	if err != nil {
		return err
	}

	if definition.Approved {
		return constants.ErrorDefinitionAlreadyApproved
	}

	err = ApproveSource(definition.Source, userId)

	if err != nil {
		return err
	}

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"approved_by":   userId,
			"approved_date": time.Now(),
			"approved":      true,
		},
	}

	var result bson.M
	updateError := definitionsCollection.FindOneAndUpdate(dbContext, filter, update, nil).Decode(&result)
	// TODO: send email to user

	if updateError != nil {
		if updateError == mongo.ErrNoDocuments {
			return constants.ErrorDefinitionNotFound
		}
		return updateError
	}

	return nil

}

func RejectDefinition(definitionId string, content string, userId string) error {

	definitionObjectId, err := primitive.ObjectIDFromHex(definitionId)

	if err != nil {
		return constants.ErrorInvalidID
	}

	definition, findError := GetDefinitionByObjectId(definitionObjectId)

	if findError != nil {
		return constants.ErrorDefinitionNotFound
	}

	if definition.Approved {
		return constants.ErrorDefinitionAlreadyApproved
	}

	rejection := types.Rejection{
		ID:           primitive.NewObjectID(),
		RejectedBy:   userId,
		RejectedDate: time.Now(),
		Content:      content,
	}

	var latestRejectionDate time.Time
	for _, d := range *definition.RejectionLog {
		if d.RejectedDate.After(latestRejectionDate) {
			latestRejectionDate = d.RejectedDate
		}
	}

	if !latestRejectionDate.IsZero() && latestRejectionDate.After(definition.LastChangeDate) {
		return constants.ErrorDefinitionRejectionNotAnsweredYet
	}

	filter := bson.M{"_id": definitionObjectId}

	update := bson.M{
		"$push": bson.M{
			"rejection_log": rejection,
		},
	}

	result := definitionsCollection.FindOneAndUpdate(dbContext, filter, update, nil)
	// TODO: send email to user

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return constants.ErrorDefinitionNotFound
		}
		return result.Err()
	}

	return nil

}

func ChangeDefinition(request *types.ChangeDefinitionRequest, userId string) error {

	id, err := primitive.ObjectIDFromHex(*request.ID)

	if err != nil {
		return constants.ErrorInvalidID
	}

	definition, findError := GetDefinitionByObjectId(id)

	if findError != nil {
		return constants.ErrorDefinitionNotFound
	}

	if definition.Approved {
		return constants.ErrorDefinitionAlreadyApproved
	}

	if definition.SubmittedBy != userId {
		return constants.ErrorDefinitionRejectionBelongsToAnotherUser
	}

	filter := bson.M{"_id": id}

	var updateEntries bson.D
	if request.Title != nil {
		updateEntries = append(updateEntries, bson.E{Key: "title", Value: request.Title})
	}
	if request.Content != nil {
		updateEntries = append(updateEntries, bson.E{Key: "content", Value: request.Content})
	}
	if request.SourceId != nil {

		sourceId, err := primitive.ObjectIDFromHex(*request.SourceId)

		if err != nil {
			return constants.ErrorInvalidID
		}

		sourceExistsError := validateSourceExists(sourceId)

		if sourceExistsError != nil {
			return sourceExistsError
		}

		updateEntries = append(updateEntries, bson.E{Key: "source", Value: sourceId})

	}

	if request.Category != nil {
		updateEntries = append(updateEntries, bson.E{Key: "category", Value: request.Category})
	}

	if len(updateEntries) > 0 {

		updateEntries = append(updateEntries, bson.E{Key: "last_change_date", Value: time.Now()})
		update := bson.M{"$set": updateEntries}

		result := definitionsCollection.FindOneAndUpdate(dbContext, filter, update, nil)

		if result.Err() != nil {
			if result.Err() == mongo.ErrNoDocuments {
				return constants.ErrorDefinitionNotFound
			}
			return result.Err()
		}
	}

	return nil

}

func GetDefinitionById(id string) (*types.Definition, error) {

	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, constants.ErrorInvalidID
	}

	filter := bson.M{"_id": objectId}
	return getDefinition(filter, nil)

}

func GetDefinitionByObjectId(id primitive.ObjectID) (*types.Definition, error) {

	filter := bson.M{"_id": id}
	return getDefinition(filter, nil)

}

func GetNewestDefinitions(limit int) ([]*types.Definition, error) {

	options := options.Find().SetSort(bson.M{"creation_date": -1}).SetLimit(int64(limit))
	return getDocuments[types.Definition](definitionsCollection, bson.M{"approved": true}, options)

}

func GetDefinitions(request *types.DefinitionPageRequest) ([]*types.Definition, error) {

	if request.PageSize <= 0 || request.Page <= 0 {
		return nil, constants.ErrorInvalidType
	}

	options := options.FindOptions{}

	if request.Sort != nil {
		options.SetSort(*request.Sort)
	}
	options.SetLimit(int64(request.PageSize))
	options.SetSkip(int64((request.Page - 1) * request.PageSize))

	filter := CreateDefinitonFilterQuery(request.Filter)
	fmt.Println("FILTER_QUERY")
	fmt.Println(filter)
	return getDocuments[types.Definition](definitionsCollection, filter, &options)

}

func CreateDefinitonFilterQuery(filter *types.DefinitionFilter) bson.D {

	query := bson.D{}

	if filter == nil {
		return query
	}

	textSearch := ""
	if filter.Content != nil && len(*filter.Content) > 0 {
		textSearch = *filter.Content
	}

	if filter.Content != nil && len(*filter.Content) > 0 {
		if len(textSearch) > 0 {
			textSearch += " "
		}
		textSearch += *filter.Content
	}

	if len(textSearch) > 0 {
		query = append(query, bson.E{Key: "$text", Value: bson.D{{Key: "$search", Value: textSearch}}})
	}

	if filter.Categories != nil && len(*filter.Categories) > 0 {
		query = append(query, bson.E{Key: "category", Value: bson.D{{Key: "$in", Value: *filter.Categories}}})
	}

	query = append(query, bson.E{Key: "approved", Value: filter.Approved})

	// bson.D{{Key: "title", Value: bson.D{{Key: "$regex", Value: primitive.Regex{Pattern: *filter.Title, Options: "i"}}}}}
	return query

}

func GetDefinitionCount() (int64, error) {

	filter := bson.M{
		"approved": true,
	}
	count, err := definitionsCollection.CountDocuments(dbContext, filter, nil)
	return count, err

}

func GetDefinitionCountInCurrentQuarter() (int64, error) {

	currentQuarterDate := common.GetCurrentQuarterDate()

	filter := bson.M{
		"approved": true,
		"approved_date": bson.M{
			"$gte": currentQuarterDate,
		},
	}

	count, err := definitionsCollection.CountDocuments(dbContext, filter, nil)
	return count, err

}

func CountDefinitionsWithSource(id primitive.ObjectID) (int, error) {

	filter := bson.M{
		"source": id,
	}

	return countDocuments(definitionsCollection, filter, nil)

}

func GetDefinitionsWithSource(id primitive.ObjectID) (*[]*types.Definition, error) {

	filter := bson.M{
		"source": id,
	}

	options := options.FindOptions{}
	definitions, err := getDocuments[types.Definition](definitionsCollection, filter, &options)

	if err != nil {
		return nil, err
	}

	return &definitions, nil

}

func DeleteDefinition(definitionId string) error {

	id, err := primitive.ObjectIDFromHex(definitionId)

	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": id,
	}

	result, err := definitionsCollection.DeleteOne(dbContext, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return constants.ErrorDefinitionNotFound
	}

	return nil

}

func getDefinition(filter interface{}, options *options.FindOneOptions) (*types.Definition, error) {

	var definition types.Definition
	err := definitionsCollection.FindOne(dbContext, filter, options).Decode(&definition)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, constants.ErrorDefinitionNotFound
		}
		return nil, err
	}

	return &definition, nil
}

func GetDefinitionPageCount(request *types.DefinitionPageCountRequest) (int64, error) {
	filter := CreateDefinitonFilterQuery(request.Filter)
	return getPageCount(definitionsCollection, request.PageSize, filter)
}
