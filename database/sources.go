package database

import (
	"fmt"
	"time"
	"yacoid_server/common"
	"yacoid_server/constants"
	"yacoid_server/types"

	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateSource(request *types.CreateSourceRequest, userId string) (*primitive.ObjectID, error) {

	var source types.Source

	now := time.Now()
	source.ID = primitive.NewObjectID()
	source.SubmittedBy = userId
	source.SubmittedDate = now
	source.LastChangeDate = now
	source.ApprovedBy = nil
	source.ApprovedDate = nil
	source.Approved = false

	source.Title = request.Title
	source.Type = request.Type

	if request.Type == types.EnumSourceType.Book && request.BookProperties != nil {
		source.BookProperties = request.BookProperties
	} else if request.Type == types.EnumSourceType.Journal && request.JournalProperties != nil {
		source.JournalProperties = request.JournalProperties
	} else if request.Type == types.EnumSourceType.Web && request.WebProperties != nil {
		source.WebProperties = request.WebProperties
	} else {
		return nil, constants.ErrorSourceCreation
	}

	authorIds, err := stringsToObjectIDs(&request.Authors)

	if err != nil {
		return nil, err
	}

	err = validateAuthorsExist(&authorIds)

	if err != nil {
		return nil, err
	}

	source.Authors = authorIds

	_, err = sourcesCollection.InsertOne(dbContext, source)

	if err != nil {
		return nil, err
	}

	return &source.ID, nil

}

func DeleteSource(sourceId string) (*[]string, error) {

	id, err := primitive.ObjectIDFromHex(sourceId)

	if err != nil {
		return nil, err
	}

	definitions, err := GetDefinitionsWithSource(id)

	if err != nil {
		return nil, err
	}
	if len(*definitions) > 0 {

		definitionIds := []string{}

		for _, definition := range *definitions {
			definitionIds = append(definitionIds, definition.ID.Hex())
		}

		return &definitionIds, constants.ErrorSourceDeletionBecauseInUse

	}

	filter := bson.M{
		"_id": id,
	}

	result, err := sourcesCollection.DeleteOne(dbContext, filter)

	if err != nil {
		return nil, err
	}

	if result.DeletedCount == 0 {
		return nil, constants.ErrorSourceNotFound
	}

	return nil, nil

}

func validateSourceExists(id primitive.ObjectID) error {

	_, err := GetSource(id)
	return err

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
			return nil, constants.ErrorSourceNotFound
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

func CountSourcesWithAuthor(authorId primitive.ObjectID) (int, error) {

	authors := []primitive.ObjectID{authorId}
	filter := bson.M{
		"authors": bson.D{{Key: "$in", Value: authors}},
	}

	return countDocuments(sourcesCollection, filter, nil)

}

func GetSources(filter interface{}) (*[]*types.Source, error) {

	options := options.FindOptions{}
	sources, err := getDocuments[types.Source](sourcesCollection, filter, &options)

	if err != nil {
		return nil, err
	}

	return &sources, nil

}

func GetSourcesWithAuthor(authorId primitive.ObjectID) (*[]*types.Source, error) {

	authors := []primitive.ObjectID{authorId}
	filter := bson.M{
		"authors": bson.D{{Key: "$in", Value: authors}},
	}

	options := options.FindOptions{}
	sources, err := getDocuments[types.Source](sourcesCollection, filter, &options)

	if err != nil {
		return nil, err
	}

	return &sources, nil

}

func ApproveSource(sourceId primitive.ObjectID, userId string) error {

	source, err := GetSource(sourceId)

	if err != nil {
		return err
	}

	if source.Approved {
		return constants.ErrorSourceAlreadyApproved
	}

	err = ApproveAuthors(source.Authors, userId)

	if err != nil {
		return err
	}

	filter := bson.M{
		"_id":      sourceId,
		"approved": false,
	}

	update := bson.M{
		"$set": bson.M{
			"approved_by":   userId,
			"approved_date": time.Now(),
			"approved":      true,
		},
	}

	var result bson.M
	updateError := sourcesCollection.FindOneAndUpdate(dbContext, filter, update, nil).Decode(&result)

	if updateError != nil {
		if updateError == mongo.ErrNoDocuments {
			return constants.ErrorSourceNotFound
		}
		return updateError
	}

	return nil

}

func ChangeSource(request *types.ChangeSourceRequest, userId string, validate *validator.Validate) error {

	id, err := primitive.ObjectIDFromHex(request.ID)

	if err != nil {
		return constants.ErrorInvalidID
	}

	source, err := GetSource(id)

	if err != nil {
		return err
	}

	if source.Approved {
		return constants.ErrorSourceAlreadyApproved
	}

	oldSource := types.Source{}
	copier.Copy(oldSource, source)

	/* Change fields */

	changed := false

	if request.Type != nil {
		source.Type = *request.Type
		changed = true
	}

	if request.Title != nil {
		source.Title = *request.Title
		changed = true
	}

	if request.Authors != nil {

		authorIds, err := stringsToObjectIDs(request.Authors)

		if err != nil {
			return err
		}

		err = validateAuthorsExist(&authorIds)

		if err != nil {
			return err
		}

		source.Authors = authorIds
		changed = true

	}

	if source.Type == types.EnumSourceType.Book && request.BookProperties != nil {

		if source.BookProperties == nil {
			source.BookProperties = &types.BookProperties{}
		}

		if request.BookProperties.PublicationDate != nil {
			fmt.Printf("Changed \"PublicationDate\" from \"%s\" to \"%s\"\n", source.BookProperties.PublicationDate, *request.BookProperties.PublicationDate)
			source.BookProperties.PublicationDate = *request.BookProperties.PublicationDate
			changed = true
		}

		if request.BookProperties.PublicationPlace != nil {
			fmt.Printf("Changed \"PublicationPlace\" from \"%s\" to \"%s\"\n", source.BookProperties.PublicationPlace, *request.BookProperties.PublicationPlace)
			source.BookProperties.PublicationPlace = *request.BookProperties.PublicationPlace
			changed = true
		}

		if request.BookProperties.PagesFrom != nil {
			fmt.Printf("Changed \"PagesFrom\" from \"%v\" to \"%v\"\n", source.BookProperties.PagesFrom, *request.BookProperties.PagesFrom)
			source.BookProperties.PagesFrom = *request.BookProperties.PagesFrom
			changed = true
		}

		if request.BookProperties.PagesTo != nil {
			fmt.Printf("Changed \"PagesTo\" from \"%v\" to \"%v\"\n", source.BookProperties.PagesTo, *request.BookProperties.PagesTo)
			source.BookProperties.PagesTo = *request.BookProperties.PagesTo
			changed = true
		}

		if request.BookProperties.Edition != nil {
			fmt.Printf("Changed \"Edition\" from \"%s\" to \"%s\"\n", source.BookProperties.Edition, *request.BookProperties.Edition)
			source.BookProperties.Edition = *request.BookProperties.Edition
			changed = true
		}

		if request.BookProperties.Publisher != nil {
			fmt.Printf("Changed \"Publisher\" from \"%s\" to \"%s\"\n", source.BookProperties.Publisher, *request.BookProperties.Publisher)
			source.BookProperties.Publisher = *request.BookProperties.Publisher
			changed = true
		}

		if request.BookProperties.ISBN != nil {
			fmt.Printf("Changed \"ISBN\" from \"%s\" to \"%s\"\n", source.BookProperties.ISBN, *request.BookProperties.ISBN)
			source.BookProperties.ISBN = *request.BookProperties.ISBN
			changed = true
		}

		if request.BookProperties.EAN != nil {
			fmt.Printf("Changed \"EAN\" from \"%s\" to \"%s\"\n", source.BookProperties.EAN, *request.BookProperties.EAN)
			source.BookProperties.EAN = *request.BookProperties.EAN
			changed = true
		}

		if request.BookProperties.DOI != nil {
			fmt.Printf("Changed \"DOI\" from \"%s\" to \"%s\"\n", source.BookProperties.DOI, *request.BookProperties.DOI)
			source.BookProperties.DOI = *request.BookProperties.DOI
			changed = true
		}

		if request.BookProperties.WebProperties != nil {

			if source.BookProperties.WebProperties == nil {
				source.BookProperties.WebProperties = &types.WebProperties{}
			}

			if request.BookProperties.WebProperties.URL != nil {
				fmt.Printf("Changed \"WebProperties.URL\" from \"%s\" to \"%s\"\n", source.BookProperties.WebProperties.URL, *request.BookProperties.WebProperties.URL)
				source.BookProperties.WebProperties.URL = *request.BookProperties.WebProperties.URL
				changed = true
			}

			if request.BookProperties.WebProperties.WebsiteName != nil {
				fmt.Printf("Changed \"WebProperties.WebsiteName\" from \"%s\" to \"%s\"\n", source.BookProperties.WebProperties.WebsiteName, *request.BookProperties.WebProperties.WebsiteName)
				source.BookProperties.WebProperties.WebsiteName = *request.BookProperties.WebProperties.WebsiteName
				changed = true
			}

			if request.BookProperties.WebProperties.AccessDate != nil {
				fmt.Printf("Changed \"WebProperties.AccessDate\" from \"%s\" to \"%s\"\n", source.BookProperties.WebProperties.AccessDate, *request.BookProperties.WebProperties.AccessDate)
				source.BookProperties.WebProperties.AccessDate = *request.BookProperties.WebProperties.AccessDate
				changed = true
			}

			if request.BookProperties.WebProperties.PublicationDate != nil {
				fmt.Printf("Changed \"WebProperties.PublicationDate\" from \"%s\" to \"%s\"\n", source.BookProperties.WebProperties.PublicationDate, *request.BookProperties.WebProperties.PublicationDate)
				source.BookProperties.WebProperties.PublicationDate = *request.BookProperties.WebProperties.PublicationDate
				changed = true
			}

		}

		/* Remove old data in other properties */
		if changed == true {
			source.JournalProperties = nil
			source.WebProperties = nil
		}

	}

	if source.Type == types.EnumSourceType.Journal && request.JournalProperties != nil {

		if source.JournalProperties == nil {
			source.JournalProperties = &types.JournalProperties{}
		}

		if request.JournalProperties.PublicationDate != nil {
			fmt.Printf("Changed \"PublicationDate\" from \"%s\" to \"%s\"\n", source.JournalProperties.PublicationDate, *request.JournalProperties.PublicationDate)
			source.JournalProperties.PublicationDate = *request.JournalProperties.PublicationDate
			changed = true
		}

		if request.JournalProperties.PublicationPlace != nil {
			fmt.Printf("Changed \"PublicationPlace\" from \"%s\" to \"%s\"\n", source.JournalProperties.PublicationPlace, *request.JournalProperties.PublicationPlace)
			source.JournalProperties.PublicationPlace = *request.JournalProperties.PublicationPlace
			changed = true
		}

		if request.JournalProperties.PagesFrom != nil {
			fmt.Printf("Changed \"PagesFrom\" from \"%v\" to \"%v\"\n", source.JournalProperties.PagesFrom, *request.JournalProperties.PagesFrom)
			source.JournalProperties.PagesFrom = *request.JournalProperties.PagesFrom
			changed = true
		}

		if request.JournalProperties.PagesTo != nil {
			fmt.Printf("Changed \"PagesTo\" from \"%v\" to \"%v\"\n", source.JournalProperties.PagesTo, *request.JournalProperties.PagesTo)
			source.JournalProperties.PagesTo = *request.JournalProperties.PagesTo
			changed = true
		}

		if request.JournalProperties.Edition != nil {
			fmt.Printf("Changed \"Edition\" from \"%s\" to \"%s\"\n", source.JournalProperties.Edition, *request.JournalProperties.Edition)
			source.JournalProperties.Edition = *request.JournalProperties.Edition
			changed = true
		}

		if request.JournalProperties.Publisher != nil {
			fmt.Printf("Changed \"Publisher\" from \"%s\" to \"%s\"\n", source.JournalProperties.Publisher, *request.JournalProperties.Publisher)
			source.JournalProperties.Publisher = *request.JournalProperties.Publisher
			changed = true
		}

		if request.JournalProperties.DOI != nil {
			fmt.Printf("Changed \"DOI\" from \"%s\" to \"%s\"\n", source.JournalProperties.DOI, *request.JournalProperties.DOI)
			source.JournalProperties.DOI = *request.JournalProperties.DOI
			changed = true
		}

		if request.JournalProperties.JournalName != nil {
			fmt.Printf("Changed \"JournalName\" from \"%s\" to \"%s\"\n", source.JournalProperties.JournalName, *request.JournalProperties.JournalName)
			source.JournalProperties.JournalName = *request.JournalProperties.JournalName
			changed = true
		}

		if request.JournalProperties.WebProperties != nil {

			if source.JournalProperties.WebProperties == nil {
				source.JournalProperties.WebProperties = &types.WebProperties{}
			}

			if request.JournalProperties.WebProperties.URL != nil {
				fmt.Printf("Changed \"WebProperties.URL\" from \"%s\" to \"%s\"\n", source.JournalProperties.WebProperties.URL, *request.JournalProperties.WebProperties.URL)
				source.JournalProperties.WebProperties.URL = *request.JournalProperties.WebProperties.URL
				changed = true
			}

			if request.JournalProperties.WebProperties.WebsiteName != nil {
				fmt.Printf("Changed \"WebProperties.WebsiteName\" from \"%s\" to \"%s\"\n", source.JournalProperties.WebProperties.WebsiteName, *request.JournalProperties.WebProperties.WebsiteName)
				source.JournalProperties.WebProperties.WebsiteName = *request.JournalProperties.WebProperties.WebsiteName
				changed = true
			}

			if request.JournalProperties.WebProperties.AccessDate != nil {
				fmt.Printf("Changed \"WebProperties.AccessDate\" from \"%s\" to \"%s\"\n", source.JournalProperties.WebProperties.AccessDate, *request.JournalProperties.WebProperties.AccessDate)
				source.JournalProperties.WebProperties.AccessDate = *request.JournalProperties.WebProperties.AccessDate
				changed = true
			}

			if request.JournalProperties.WebProperties.PublicationDate != nil {
				fmt.Printf("Changed \"WebProperties.PublicationDate\" from \"%s\" to \"%s\"\n", source.JournalProperties.WebProperties.PublicationDate, *request.JournalProperties.WebProperties.PublicationDate)
				source.JournalProperties.WebProperties.PublicationDate = *request.JournalProperties.WebProperties.PublicationDate
				changed = true
			}

		}

		/* Remove old data in other properties */
		if changed == true {
			source.BookProperties = nil
			source.WebProperties = nil
		}

	}

	if source.Type == types.EnumSourceType.Web && request.WebProperties != nil {

		if source.WebProperties == nil {
			source.WebProperties = &types.WebProperties{}
		}

		if request.WebProperties.URL != nil {
			fmt.Printf("Changed \"URL\" from \"%s\" to \"%s\"\n", source.WebProperties.URL, *request.WebProperties.URL)
			source.WebProperties.URL = *request.WebProperties.URL
			changed = true
		}

		if request.WebProperties.WebsiteName != nil {
			fmt.Printf("Changed \"WebsiteName\" from \"%s\" to \"%s\"\n", source.WebProperties.WebsiteName, *request.WebProperties.WebsiteName)
			source.WebProperties.WebsiteName = *request.WebProperties.WebsiteName
			changed = true
		}

		if request.WebProperties.AccessDate != nil {
			fmt.Printf("Changed \"AccessDate\" from \"%s\" to \"%s\"\n", source.WebProperties.AccessDate, *request.WebProperties.AccessDate)
			source.WebProperties.AccessDate = *request.WebProperties.AccessDate
			changed = true
		}

		if request.WebProperties.PublicationDate != nil {
			fmt.Printf("Changed \"PublicationDate\" from \"%s\" to \"%s\"\n", source.WebProperties.PublicationDate, *request.WebProperties.PublicationDate)
			source.WebProperties.PublicationDate = *request.WebProperties.PublicationDate
			changed = true
		}

		/* Remove old data in other properties */
		if changed == true {
			source.BookProperties = nil
			source.JournalProperties = nil
		}

	}

	/* Validate */
	errorFields := source.Validate(validate)

	if errorFields != nil {
		return constants.CreateValidationError(&errorFields)
	}

	/* Update existing source */

	source.LastChangeDate = time.Now()

	filter := bson.M{
		"_id": id,
	}

	result, err := sourcesCollection.ReplaceOne(dbContext, filter, source, nil)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return constants.ErrorSourceNotFound
	}

	return nil

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
