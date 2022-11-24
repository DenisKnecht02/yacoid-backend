package database

import (
	"errors"
	"fmt"
	"time"

	"yacoid_server/common"
	"yacoid_server/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrorDefinitionAlreadyApproved = errors.New("DEFINITION_ALREADY_APPROVED")
var ErrorDefinitionRejectionNotAnsweredYet = errors.New("DEFINITION_REJECTION_NOT_ANSWERED_YET")
var ErrorDefinitionRejectionBelongsToAnotherUser = errors.New("DEFINITION_REJECTION_BELONGS_TO_ANOTHER_USER")

type Rejection struct {
	ID           primitive.ObjectID `bson:"_id" json:"-"`
	RejectedBy   string             `bson:"rejected_by" json:"rejectedBy" validate:"required"`
	RejectedDate time.Time          `bson:"rejected_date" json:"rejectedDate" validate:"required"`
	Content      string             `bson:"content" json:"content" validate:"required"`
}

type Definition struct {
	ID                   primitive.ObjectID `bson:"_id" json:"id"`
	SubmittedBy          string             `bson:"submitted_by" json:"submittedBy"`
	SubmittedDate        time.Time          `bson:"submitted_date" json:"submittedDate"`
	LastSubmitChangeDate time.Time          `bson:"last_submit_change_date" json:"lastSubmitChangeDate"`
	ApprovedBy           *string            `bson:"approved_by" json:"approvedBy"`
	ApprovedDate         *time.Time         `bson:"approved_date" json:"approvedDate"`
	Approved             bool               `bson:"approved" json:"approved"`
	RejectionLog         *[]*Rejection      `bson:"rejection_log" json:"-"`
	Title                string             `bson:"title" json:"title"`
	Content              string             `bson:"content" json:"content"`
	Source               primitive.ObjectID `bson:"source" json:"source"`
	PublishingDate       time.Time          `bson:"publishing_date" json:"publishingDate"`
	Tags                 *[]string          `bson:"tags" json:"tags"`
}

func (definition *Definition) IsApproved() bool {
	return definition.ApprovedBy != nil && definition.ApprovedDate != nil
}

func SubmitDefinition(request *types.SubmitDefinitionRequest, userId string) (*Definition, error) {

	var definition Definition

	now := time.Now()
	definition.ID = primitive.NewObjectID()
	definition.SubmittedBy = userId
	definition.SubmittedDate = now
	definition.LastSubmitChangeDate = now
	definition.ApprovedBy = nil
	definition.ApprovedDate = nil
	definition.Approved = false

	definition.Title = request.Title
	definition.Content = request.Content
	definition.Tags = request.Tags
	definition.PublishingDate = request.PublishingDate

	sourceId, sourceIdError := primitive.ObjectIDFromHex(request.Source)

	if sourceIdError != nil {
		return nil, sourceIdError
	}

	sourceExistsError := validateSourceExists(sourceId)

	if sourceExistsError != nil {
		return nil, sourceExistsError
	}

	definition.Source = sourceId

	rejectionLog := []*Rejection{}
	definition.RejectionLog = &rejectionLog

	if definition.Tags == nil {
		definition.Tags = &[]string{}
	}

	_, err := definitionsCollection.InsertOne(dbContext, definition)
	// TODO: send email to user??

	if err != nil {
		return nil, err
	}

	return &definition, nil

}

func validateSourceExists(id primitive.ObjectID) error {

	_, err := GetSource(id)
	return err

}

func ApproveDefinition(definitionId string, userId string) error {

	definitionObjectId, definitionObjectIdError := primitive.ObjectIDFromHex(definitionId)

	if definitionObjectIdError != nil {
		return InvalidID
	}

	filter := bson.M{"_id": definitionObjectId}
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
			return ErrorDefinitionNotFound
		}
		return updateError
	}

	return nil

}

func RejectDefinition(definitionId string, content string, userId string) error {

	definitionObjectId, definitionObjectIdError := primitive.ObjectIDFromHex(definitionId)

	if definitionObjectIdError != nil {
		return InvalidID
	}

	definition, findError := GetDefinitionByObjectId(definitionObjectId)

	if findError != nil {
		return ErrorDefinitionNotFound
	}

	if definition.Approved == true {
		return ErrorDefinitionAlreadyApproved
	}

	rejection := Rejection{
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

	if !latestRejectionDate.IsZero() && latestRejectionDate.After(definition.LastSubmitChangeDate) {
		return ErrorDefinitionRejectionNotAnsweredYet
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
			return ErrorDefinitionNotFound
		}
		return result.Err()
	}

	return nil

}

func ChangeDefinition(definitionId string, title *string, content *string, source *types.Source, tags *[]string, userId string) error {

	definitionObjectId, definitionObjectIdError := primitive.ObjectIDFromHex(definitionId)

	if definitionObjectIdError != nil {
		return InvalidID
	}

	definition, findError := GetDefinitionByObjectId(definitionObjectId)

	if findError != nil {
		return ErrorDefinitionNotFound
	}

	if definition.Approved == true {
		return ErrorDefinitionAlreadyApproved
	}

	if definition.SubmittedBy != userId {
		return ErrorDefinitionRejectionBelongsToAnotherUser
	}

	filter := bson.M{"_id": definitionObjectId}

	var updateEntries bson.D
	if title != nil {
		updateEntries = append(updateEntries, bson.E{Key: "title", Value: title})
	}
	if content != nil {
		updateEntries = append(updateEntries, bson.E{Key: "content", Value: content})
	}
	if source != nil {
		updateEntries = append(updateEntries, bson.E{Key: "source", Value: source})
	}
	if tags != nil {
		updateEntries = append(updateEntries, bson.E{Key: "tags", Value: tags})
	}

	if len(updateEntries) > 0 {

		updateEntries = append(updateEntries, bson.E{Key: "last_submit_change_date", Value: time.Now()})
		update := bson.M{"$set": updateEntries}

		result := definitionsCollection.FindOneAndUpdate(dbContext, filter, update, nil)

		if result.Err() != nil {
			if result.Err() == mongo.ErrNoDocuments {
				return ErrorDefinitionNotFound
			}
			return result.Err()
		}
	}

	return nil

}

func GetDefinitionById(id string) (*Definition, error) {

	objectId, idError := primitive.ObjectIDFromHex(id)

	if idError != nil {
		return nil, InvalidID
	}

	filter := bson.M{"_id": objectId}
	return getDefinition(filter, nil)

}

func GetDefinitionByObjectId(id primitive.ObjectID) (*Definition, error) {

	filter := bson.M{"_id": id}
	return getDefinition(filter, nil)

}

func GetNewestDefinitions(limit int) ([]*Definition, error) {

	options := options.Find().SetSort(bson.M{"creation_date": -1}).SetLimit(int64(limit))
	return getDocuments[Definition](definitionsCollection, bson.M{"approved": true}, options)

}

func GetDefinitions(pageSize int, page int, definitionFilter *types.DefinitionFilter, sort *interface{}) ([]*Definition, error) {

	if pageSize <= 0 || page <= 0 {
		return nil, common.ErrorInvalidType
	}

	options := options.FindOptions{}

	if sort != nil {
		options.SetSort(*sort)
	}
	options.SetLimit(int64(pageSize))
	options.SetSkip(int64((page - 1) * pageSize))

	filter := CreateDefinitonFilterQuery(definitionFilter)
	fmt.Println("FILTER_QUERY")
	fmt.Println(filter)
	return getDocuments[Definition](definitionsCollection, filter, &options)

}

func CreateDefinitonFilterQuery(filter *types.DefinitionFilter) bson.D {

	query := bson.D{}

	if filter == nil {
		return query
	}

	textSearch := ""
	if filter.Title != nil && len(*filter.Title) > 0 {
		textSearch = *filter.Title
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

	if filter.Tags != nil {
		query = append(query, bson.E{Key: "tags", Value: bson.D{{Key: "$in", Value: *filter.Tags}}})
	}

	// TODO: Sources, Authors, PublishingDates

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

func getDefinition(filter interface{}, options *options.FindOneOptions) (*Definition, error) {

	var definition Definition
	err := definitionsCollection.FindOne(dbContext, filter, options).Decode(&definition)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrorDefinitionNotFound
		}
		return nil, err
	}

	return &definition, nil
}

func GetDefinitionPageCount(request *types.DefinitionPageCountRequest) (int64, error) {
	filter := CreateDefinitonFilterQuery(request.Filter)
	return getPageCount(definitionsCollection, request.PageSize, filter)
}
