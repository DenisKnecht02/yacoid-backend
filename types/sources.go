package types

import (
	"strings"
	"time"
	"yacoid_server/common"
	"yacoid_server/constants"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Source struct {
	ID                primitive.ObjectID   `bson:"_id" json:"-"`
	SubmittedBy       string               `bson:"submitted_by" json:"submittedBy"`
	SubmittedDate     time.Time            `bson:"submitted_date" json:"submittedDate"`
	LastChangeDate    time.Time            `bson:"last_change_date" json:"lastChangeDate"`
	ApprovedBy        *string              `bson:"approved_by" json:"approvedBy"`
	ApprovedDate      *time.Time           `bson:"approved_date" json:"approvedDate"`
	Approved          bool                 `bson:"approved" json:"approved"`
	Type              SourceType           `bson:"type" json:"type" validate:"required,is-source-type"`
	Authors           []primitive.ObjectID `bson:"authors" json:"authors" validate:"required,min=1"`
	BookProperties    *BookProperties      `bson:"book_properties" json:"bookProperties" validate:"required_without_all=JournalProperties WebProperties,omitempty,dive"`
	JournalProperties *JournalProperties   `bson:"journal_properties" json:"journalProperties" validate:"required_without_all=BookProperties WebProperties,omitempty,dive"`
	WebProperties     *WebProperties       `bson:"web_properties" json:"webProperties" validate:"required_without_all=BookProperties JournalProperties,omitempty,dive"`
}

func (object *Source) Validate(validate *validator.Validate) []string {

	errorFields := common.ValidateStruct(object, validate)

	if object.Type == EnumSourceType.Book && object.BookProperties == nil {
		errorFields = append(errorFields, "BookProperties missing")
	} else if object.Type == EnumSourceType.Journal && object.JournalProperties == nil {
		errorFields = append(errorFields, "JournalProperties missing")
	} else if object.Type == EnumSourceType.Web && object.WebProperties == nil {
		errorFields = append(errorFields, "WebProperties missing")
	}

	return errorFields

}

type BookProperties struct {
	Title            string    `bson:"title" json:"title" validate:"required,min=1"`
	PublicationDate  time.Time `bson:"publication_date" json:"publicationDate" validate:"omitempty"`
	PublicationPlace string    `bson:"publication_place" json:"publicationPlace" validate:"omitempty"`
	PagesFrom        int       `bson:"pages_from" json:"pagesFrom" validate:"omitempty,min=1"`
	PagesTo          int       `bson:"pages_to" json:"pagesTo" validate:"omitempty,min=1"`
	Edition          string    `bson:"edition" json:"edition" validate:"omitempty,min=1"`
	Publisher        string    `bson:"publisher" json:"publisher" validate:"omitempty,min=1"`
	ISBN             string    `bson:"isbn" json:"isbn" validate:"omitempty,isbn"`
	EAN              string    `bson:"ean" json:"ean" validate:"omitempty,min=1"`
	DOI              string    `bson:"doi" json:"doi" validate:"omitempty,min=1"`
}

type JournalProperties struct {
	Title            string    `bson:"title" json:"title" validate:"required,min=1"`
	PublicationDate  time.Time `bson:"publication_date" json:"publicationDate" validate:"omitempty"`
	PublicationPlace string    `bson:"publication_place" json:"publicationPlace" validate:"omitempty"`
	PagesFrom        int       `bson:"pages_from" json:"pagesFrom" validate:"omitempty,min=1"`
	PagesTo          int       `bson:"pages_to" json:"pagesTo" validate:"omitempty,min=1"`
	DOI              string    `bson:"doi" json:"doi" validate:"omitempty,min=1"`
	JournalName      string    `bson:"journal_name" json:"journalName" validate:"required,min=1"`
	Edition          string    `bson:"edition" json:"edition" validate:"omitempty,min=1"`
	Publisher        string    `bson:"publisher" json:"publisher" validate:"omitempty,min=1"`
}

type WebProperties struct {
	ArticleName     string    `bson:"article_name" json:"articleName" validate:"required,min=1"`
	URL             string    `bson:"url" json:"url" validate:"required,url"`
	WebsiteName     string    `bson:"website_name" json:"websiteName" validate:"required,min=1"`
	AccessDate      time.Time `bson:"access_date" json:"accessDate" validate:"required"`
	PublicationDate time.Time `bson:"publication_date" json:"publicationDate" validate:"omitempty"`
}

type CreateSourceRequest struct {
	Type              SourceType         `bson:"type" json:"type" validate:"required,is-source-type"`
	Authors           []string           `bson:"authors" json:"authors" validate:"required,min=1"`
	Title             string             `bson:"title" json:"title" validate:"required,min=1"`
	BookProperties    *BookProperties    `bson:"book_properties" json:"bookProperties" validate:"required_without_all=JournalProperties WebProperties,omitempty,dive"`
	JournalProperties *JournalProperties `bson:"journal_properties" json:"journalProperties" validate:"required_without_all=BookProperties WebProperties,omitempty,dive"`
	WebProperties     *WebProperties     `bson:"web_properties" json:"webProperties" validate:"required_without_all=BookProperties JournalProperties,omitempty,dive"`
}

func (object *CreateSourceRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(object, validate)
}

type ChangeSourceRequest struct {
	ID                string                   `json:"id" validate:"required"`
	Type              *SourceType              `json:"type" validate:"omitempty,is-source-type"`
	Authors           *[]string                `json:"authors" validate:"omitempty,min=1"`
	BookProperties    *ChangeBookProperties    `json:"bookProperties" validate:"omitempty,dive"`
	JournalProperties *ChangeJournalProperties `json:"journalProperties" validate:"omitempty,dive"`
	WebProperties     *ChangeWebProperties     `json:"webProperties" validate:"omitempty,dive"`
}

func (object *ChangeSourceRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(object, validate)
}

type ChangeBookProperties struct {
	Title            *string    `bson:"title" json:"title" validate:"required,min=1"`
	PublicationDate  *time.Time `bson:"publication_date" json:"publicationDate" validate:"omitempty"`
	PublicationPlace *string    `bson:"publication_place" json:"publicationPlace" validate:"omitempty"`
	PagesFrom        *int       `bson:"pages_from" json:"pagesFrom" validate:"omitempty,min=1"`
	PagesTo          *int       `bson:"pages_to" json:"pagesTo" validate:"omitempty,min=1"`
	Edition          *string    `bson:"edition" json:"edition" validate:"omitempty,min=1"`
	Publisher        *string    `bson:"publisher" json:"publisher" validate:"omitempty,min=1"`
	ISBN             *string    `bson:"isbn" json:"isbn" validate:"omitempty,isbn"`
	EAN              *string    `bson:"ean" json:"ean" validate:"omitempty,min=1"`
	DOI              *string    `bson:"doi" json:"doi" validate:"omitempty,min=1"`
}

type ChangeJournalProperties struct {
	Title            *string    `bson:"title" json:"title" validate:"required,min=1"`
	PublicationDate  *time.Time `bson:"publication_date" json:"publicationDate" validate:"omitempty"`
	PublicationPlace *string    `bson:"publication_place" json:"publicationPlace" validate:"omitempty"`
	PagesFrom        *int       `bson:"pages_from" json:"pagesFrom" validate:"omitempty,min=1"`
	PagesTo          *int       `bson:"pages_to" json:"pagesTo" validate:"omitempty,min=1"`
	DOI              *string    `bson:"doi" json:"doi" validate:"omitempty,min=1"`
	JournalName      *string    `bson:"journal_name" json:"journalName" validate:"required,min=1"`
	Edition          *string    `bson:"edition" json:"edition" validate:"omitempty,min=1"`
	Publisher        *string    `bson:"publisher" json:"publisher" validate:"omitempty,min=1"`
}

type ChangeWebProperties struct {
	ArticleName     *string    `bson:"article_name" json:"articleName" validate:"required,min=1"`
	URL             *string    `bson:"url" json:"url" validate:"omitempty,url"`
	WebsiteName     *string    `bson:"website_name" json:"websiteName" validate:"omitempty,min=1"`
	AccessDate      *time.Time `bson:"access_date" json:"accessDate" validate:"omitempty"`
	PublicationDate *time.Time `bson:"publication_date" json:"publicationDate" validate:"omitempty"`
}

type SourceType string

type sourceTypeList struct {
	Unknown SourceType
	Book    SourceType
	Journal SourceType
	Web     SourceType
}

var EnumSourceType = &sourceTypeList{
	Unknown: "unknown",
	Book:    "book",
	Journal: "journal",
	Web:     "web",
}

var sourceTypeMap = map[string]SourceType{
	"book":    EnumSourceType.Book,
	"journal": EnumSourceType.Journal,
	"web":     EnumSourceType.Web,
}

func ParseStringToSourceType(str string) (SourceType, error) {
	sourceType, ok := sourceTypeMap[strings.ToLower(str)]
	if ok {
		return sourceType, nil
	} else {
		return sourceType, constants.ErrorInvalidEnum
	}
}

func (sourceType SourceType) String() string {
	switch sourceType {
	case EnumSourceType.Book:
		return "book"
	case EnumSourceType.Journal:
		return "journal"
	case EnumSourceType.Web:
		return "web"
	}
	return "unknown"
}

type SourcePageCountRequest struct {
	PageSize int           `json:"pageSize" validate:"required,min=1"`
	Filter   *SourceFilter `json:"filter" validate:"omitempty,dive"`
}

func (request *SourcePageCountRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(request, validate)
}

type SourcePageRequest struct {
	PageSize int           `json:"pageSize" validate:"required,min=1"`
	Page     int           `json:"page" validate:"required,min=1"`
	Filter   *SourceFilter `json:"filter" validate:"omitempty,dive"`
	Sort     *interface{}  `json:"sort"`
}

func (request *SourcePageRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(request, validate)
}

type SourceFilter struct {
	Approved  bool          `json:"approved" bson:"approved" validate:"omitempty"`
	Types     *[]SourceType `json:"types" bson:"types" validate:"omitempty,dive,is-source-type"`
	Title     *string       `json:"title" bson:"title" validate:"omitempty,min=1"`
	AuthorIds *[]string     `json:"authors" bson:"authors" validate:"omitempty,min=1"`
}
