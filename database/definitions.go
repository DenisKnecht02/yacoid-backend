package database

import (
	"math"
	"time"

	"yacoid_server/auth"
	"yacoid_server/common"
	"yacoid_server/constants"
	"yacoid_server/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DefinitionToUserResponse(definition *types.Definition) (*types.DefinitionsOfUserResponse, error) {

	response := types.DefinitionsOfUserResponse{}

	response.ID = definition.ID

	response.SubmittedBy = definition.SubmittedBy
	nickname, err := auth.GetNicknameOfUser(definition.SubmittedBy)

	if err == nil {
		response.SubmittedByName = nickname
	} else {
		response.SubmittedByName = "<deleted>"
	}

	response.SubmittedDate = definition.SubmittedDate

	response.ApprovedBy = definition.ApprovedBy
	response.ApprovedDate = definition.ApprovedDate
	response.Approved = definition.Approved

	response.RejectionLog = RejectionsToResponses(definition.RejectionLog)
	response.Content = definition.Content

	source, err := GetSource(definition.Source)

	if err != nil {
		return nil, err
	}

	sourceResponse, err := SourceToResponse(source)

	if err != nil {
		return nil, err
	}

	response.Source = *sourceResponse
	response.Category = definition.Category

	response.Status = definition.GetStatus()

	return &response, nil

}

func DefinitionsToUserResponses(definitions *[]*types.Definition) (*[]types.DefinitionsOfUserResponse, error) {

	responses := []types.DefinitionsOfUserResponse{}

	for _, definition := range *definitions {

		response, err := DefinitionToUserResponse(definition)

		if err != nil {
			return nil, err
		}

		responses = append(responses, *response)
	}

	return &responses, nil

}

func DefinitionToResponse(definition *types.Definition) (*types.DefinitionResponse, error) {

	response := types.DefinitionResponse{}

	response.ID = definition.ID

	response.SubmittedBy = definition.SubmittedBy
	nickname, err := auth.GetNicknameOfUser(definition.SubmittedBy)

	if err == nil {
		response.SubmittedByName = nickname
	} else {
		response.SubmittedByName = "<deleted>"
	}

	response.SubmittedDate = definition.SubmittedDate
	response.Content = definition.Content

	source, err := GetSource(definition.Source)

	if err != nil {
		return nil, err
	}

	sourceResponse, err := SourceToResponse(source)

	if err != nil {
		return nil, err
	}

	response.Source = *sourceResponse
	response.Category = definition.Category

	return &response, nil

}

func DefinitionsToResponses(definitions *[]*types.Definition) (*[]types.DefinitionResponse, error) {

	responses := []types.DefinitionResponse{}

	for _, definition := range *definitions {

		response, err := DefinitionToResponse(definition)

		if err != nil {
			return nil, err
		}

		responses = append(responses, *response)
	}

	return &responses, nil

}

func RejectionToResponse(rejection *types.Rejection) *types.RejectionResponse {

	response := types.RejectionResponse{}

	response.ID = rejection.ID

	response.RejectedBy = rejection.RejectedBy
	nickname, err := auth.GetNicknameOfUser(rejection.RejectedBy)

	if err == nil {
		response.RejectedByName = nickname
	} else {
		response.RejectedByName = "<deleted>"
	}

	response.RejectedDate = rejection.RejectedDate

	response.Content = rejection.Content

	return &response

}

func RejectionsToResponses(rejections *[]*types.Rejection) *[]*types.RejectionResponse {

	responses := []*types.RejectionResponse{}

	for _, rejection := range *rejections {

		response := RejectionToResponse(rejection)
		responses = append(responses, response)
	}

	return &responses

}

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

	definition.Content = request.Content
	definition.Category = request.Category

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

	if err != nil && err != constants.ErrorSourceAlreadyApproved {
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

	latestRejectionDate := definition.GetLatestRejection()

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
		return constants.ErrorDefinitionBelongsToAnotherUser
	}

	filter := bson.M{"_id": id}

	var updateEntries bson.D

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

	options := options.AggregateOptions{}

	filter, err := CreateDefinitonFilterQuery(int64(request.Page), int64(request.PageSize), request.Filter)

	if err != nil {
		return nil, err
	}

	return aggregateDocuments[types.Definition](definitionsCollection, *filter, &options)

}

func CreateDefinitonFilterQuery(page int64, pageSize int64, filter *types.DefinitionFilter) (*bson.A, error) {

	pipeline := bson.A{}

	if filter == nil {
		return &pipeline, nil
	}

	matchStage := bson.D{}

	textSearch := ""
	if filter.Content != nil && len(*filter.Content) > 0 {
		textSearch = *filter.Content
	}

	if len(textSearch) > 0 {
		matchStage = append(matchStage, bson.E{Key: "$text", Value: bson.M{"$search": textSearch}})
	}

	if filter.Categories != nil && len(*filter.Categories) > 0 {
		matchStage = append(matchStage, bson.E{Key: "category", Value: bson.M{"$in": *filter.Categories}})
	}

	if filter.Approved != nil {
		matchStage = append(matchStage, bson.E{Key: "approved", Value: *filter.Approved})
	}

	if filter.UserId != nil && len(*filter.UserId) > 0 {
		matchStage = append(matchStage, bson.E{Key: "submitted_by", Value: *filter.UserId})
	}

	pipeline = append(pipeline, bson.D{
		{Key: "$match", Value: matchStage},
	})

	authors, err := stringsToObjectIDs(filter.AuthorIds)
	if filter.AuthorIds != nil && len(*filter.AuthorIds) > 0 {

		if err != nil {
			return nil, err
		}

		// If we lookup on the source, the orignal source (which is just the ID) will be replaced
		// with the entire document. To avoid this we will save the original value here and
		// replace with after lookup.
		pipeline = append(pipeline, bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "original_source", Value: "$source"},
			}},
		})

		lookup := bson.D{
			{Key: "from", Value: "sources"},
			{Key: "localField", Value: "source"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "source"},
		}

		pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: lookup}})
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "source.authors", Value: bson.D{{Key: "$in", Value: authors}}}}}})

		// Replace modified source with original source
		pipeline = append(pipeline, bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "source", Value: "$original_source"},
			}},
		})

		// Remove original source
		pipeline = append(pipeline, bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "original_source", Value: 0},
			}},
		})

	}

	if page > 0 && pageSize > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$skip", Value: int64((page - 1) * pageSize)}})
	}

	if pageSize > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: int64(pageSize)}})
	}

	return &pipeline, nil

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
	filter, err := CreateDefinitionCountFilter(request.Filter)

	if err != nil {
		return 0, err
	}

	options := options.AggregateOptions{}
	response, err := aggregateDocuments[interface{}](definitionsCollection, *filter, &options)

	if err != nil {
		return 0, err
	}

	metadata := (*response[0]).(primitive.D)
	metadataContent := metadata.Map()["metadata"].(primitive.A)

	var total int32 = 0

	if len(metadataContent) > 0 {

		totalContent := metadataContent[0].(primitive.D)
		total = totalContent.Map()["total"].(int32)

	}

	return int64(math.Ceil(float64(total) / float64(request.PageSize))), nil
}

func CreateDefinitionCountFilter(filter *types.DefinitionFilter) (*bson.A, error) {

	pipeline := bson.A{}

	if filter == nil {
		return &pipeline, nil
	}

	matchStage := bson.D{}

	textSearch := ""
	if filter.Content != nil && len(*filter.Content) > 0 {
		textSearch = *filter.Content
	}

	if len(textSearch) > 0 {
		matchStage = append(matchStage, bson.E{Key: "$text", Value: bson.M{"$search": textSearch}})
	}

	if filter.Categories != nil && len(*filter.Categories) > 0 {
		matchStage = append(matchStage, bson.E{Key: "category", Value: bson.M{"$in": *filter.Categories}})
	}

	if filter.Approved != nil {
		matchStage = append(matchStage, bson.E{Key: "approved", Value: *filter.Approved})
	}

	if filter.UserId != nil && len(*filter.UserId) > 0 {
		matchStage = append(matchStage, bson.E{Key: "submitted_by", Value: *filter.UserId})
	}

	pipeline = append(pipeline, bson.D{
		{Key: "$match", Value: matchStage},
	})

	authors, err := stringsToObjectIDs(filter.AuthorIds)
	if filter.AuthorIds != nil && len(*filter.AuthorIds) > 0 {

		if err != nil {
			return nil, err
		}

		lookup := bson.D{
			{Key: "from", Value: "sources"},
			{Key: "localField", Value: "source"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "source"},
		}

		pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: lookup}})
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "source.authors", Value: bson.D{{Key: "$in", Value: authors}}}}}})

	}

	pipeline = append(pipeline, bson.D{
		{Key: "$facet",
			Value: bson.D{
				{Key: "metadata",
					Value: bson.A{
						bson.D{{Key: "$count", Value: "total"}},
					},
				},
			},
		},
	})

	return &pipeline, nil

}
