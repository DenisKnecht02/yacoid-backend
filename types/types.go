package types

import (
	"strings"
	"time"
	"yacoid_server/common"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubmitDefinitionRequest struct {
	Title          string    `json:"title" validate:"required,min=1"`
	Content        string    `json:"content" validate:"required,min=1"`
	Source         string    `json:"source" validate:"required"`
	PublishingDate time.Time `json:"publishingDate" validate:"required"`
	Tags           *[]string `json:"tags" validate:"required,min=1"`
}

func (author *SubmitDefinitionRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(author, validate)
}

type DefinitionPageRequest struct {
	PageSize int               `json:"pageSize" validate:"required"`
	Page     int               `json:"page" validate:"required,min=1"`
	Filter   *DefinitionFilter `json:"filter" validate:"omitempty,dive"`
	Sort     *interface{}      `json:"sort"`
}

func (DefinitionPageRequest *DefinitionPageRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(DefinitionPageRequest, validate)
}

type RejectRequest struct {
	ID      string `json:"id" validate:"required"`
	Content string `json:"content" validate:"required,min=1"`
}

func (rejection *RejectRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(rejection, validate)
}

type ChangeDefinitionRequest struct {
	ID             string     `json:"id" validate:"required"`
	Title          *string    `json:"title"`
	Content        *string    `json:"content"`
	Source         *Source    `json:"source" validate:"omitempty,dive"`
	PublishingDate *time.Time `json:"publishingDate" validate:"omitempty"`
	Tags           *[]string  `json:"tags"`
}

func (rejection *ChangeDefinitionRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(rejection, validate)
}

type CreateAuthorRequest struct {
	FirstName string `json:"firstName" validate:"required,min=1"`
	LastName  string `json:"lastName" validate:"required,min=1"`
}

func (rejection *CreateAuthorRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(rejection, validate)
}

type AuthorType string

type authorTypeList struct {
	Unknown      AuthorType
	Person       AuthorType
	Organization AuthorType
}

var EnumAuthorType = &authorTypeList{
	Unknown:      "unknown",
	Person:       "person",
	Organization: "organization",
}

var authorTypeMap = map[string]AuthorType{
	"person":       EnumAuthorType.Person,
	"organization": EnumAuthorType.Organization,
}

func ParseStringToAuthorType(str string) (AuthorType, error) {
	authorType, ok := authorTypeMap[strings.ToLower(str)]
	if ok {
		return authorType, nil
	} else {
		return authorType, common.ErrorInvalidEnum
	}
}

func (authorType AuthorType) String() string {
	switch authorType {
	case EnumAuthorType.Person:
		return "person"
	case EnumAuthorType.Organization:
		return "organization"
	}
	return "unknown"
}

type Author struct {
	ID            primitive.ObjectID `bson:"_id" json:"-"`
	SlugId        string             `bson:"slug_id" json:"slugId"`
	SubmittedBy   primitive.ObjectID `bson:"submitted_by" json:"submittedBy"`
	SubmittedDate time.Time          `bson:"submitted_date" json:"submittedDate"`
	Type          AuthorType         `bson:"type" json:"type"`
}

type PersonAuthor struct {
	Author
	FirstName string `bson:"first_name" json:"firstName"`
	LastName  string `bson:"last_name" json:"lastName"`
}

type OrganizationAuthor struct {
	Author
	OrganizationName string `bson:"organization_name" json:"organizationName"`
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
		return sourceType, common.ErrorInvalidEnum
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

type Source struct {
	ID            primitive.ObjectID     `bson:"_id" json:"-"`
	SubmittedBy   primitive.ObjectID     `bson:"submitted_by" json:"submittedBy"`
	SubmittedDate time.Time              `bson:"submitted_date" json:"submittedDate"`
	Type          SourceType             `bson:"type" json:"type"`
	Authors       []primitive.ObjectID   `bson:"authors" json:"authors" validate:"required,min=1"`
	Title         string                 `bson:"title" json:"title" validate:"required,min=1"`
	Properties    map[string]interface{} `bson:"properties" json:"properties" validate:"required,min=1"`
}

func (author *Source) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(author, validate)
}

type WebSourceProperties struct {
	URL             string    `bson:"url" json:"url"`
	AccessDate      time.Time `bson:"access_date" json:"accessDate"`
	PublicationDate time.Time `bson:"publication_date" json:"publicationDate"`
}

type BookSource struct {
	PublicationDate  time.Time           `bson:"publication_date" json:"publicationDate"`
	PublicationPlace string              `bson:"publication_place" json:"publicationPlace"`
	PagesFrom        int                 `bson:"pages_from" json:"pagesFrom"`
	PagesTo          int                 `bson:"pages_to" json:"pagesTo"`
	Edition          string              `bson:"edition" json:"edition"`
	Publisher        string              `bson:"publisher" json:"publisher"`
	ISBN             string              `bson:"isbn" json:"isbn"`
	EAN              string              `bson:"ean" json:"ean"`
	DOI              string              `bson:"doi" json:"doi"`
	Web              WebSourceProperties `bson:"web" json:"web"`
}

type JournalSource struct {
	PublicationDate  time.Time `bson:"publication_date" json:"publicationDate"`
	PublicationPlace string    `bson:"publication_place" json:"publicationPlace"`
	PagesFrom        int       `bson:"pages_from" json:"pagesFrom"`
	PagesTo          int       `bson:"pages_to" json:"pagesTo"`
	DOI              string    `bson:"doi" json:"doi"`
	Edition          string    `bson:"edition" json:"edition"`
	Publisher        string    `bson:"publisher" json:"publisher"`
}

type WebSource struct {
	WebSourceProperties
}

// TODO: use interfaces to add Validate method to all of them?
type CreateSourceRequest struct {
	Type       SourceType             `bson:"type" json:"type"`
	Authors    []string               `bson:"authors" json:"authors" validate:"required,min=1"`
	Title      string                 `bson:"title" json:"title" validate:"required,min=1"`
	Properties map[string]interface{} `bson:"properties" json:"properties" validate:"required"`
}

type CreateBookSourceRequest struct {
	Type    SourceType `bson:"type" json:"type"`
	Authors []string   `bson:"authors" json:"authors" validate:"required,min=1"`
	Title   string     `bson:"title" json:"title" validate:"required,min=1"`
}

type CreateJournalSourceRequest struct {
	Type    SourceType `bson:"type" json:"type"`
	Authors []string   `bson:"authors" json:"authors" validate:"required,min=1"`
	Title   string     `bson:"title" json:"title" validate:"required,min=1"`
}

func (rejection *CreateSourceRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(rejection, validate)
}

type DefinitionFilter struct {
	Title           *string      `json:"title" bson:"title" validate:"omitempty"`
	Content         *string      `json:"content" bson:"content" validate:"omitempty"`
	PublishingDates *[]time.Time `json:"publishing_dates" bson:"publishing_dates" validate:"omitempty,min=1"`
	Authors         *[]*Author   `json:"authors" bson:"authors" validate:"omitempty,min=1,dive"`
	Sources         *[]*Source   `json:"sources" bson:"sources" validate:"omitempty,min=1,dive"`
	Tags            *[]string    `json:"tags" bson:"tags" validate:"omitempty,min=1"`
}

type StatisticsResponse struct {
	DefinitionCount                 int `json:"definitionCount"`
	DefinitionCountInCurrentQuarter int `json:"definitionCountInCurrentQuarter"`
	SourceCount                     int `json:"sourceCount"`
	SourceCountInCurrentQuarter     int `json:"sourceCountInCurrentQuarter"`
	AuthorCount                     int `json:"authorCount"`
	AuthorCountInCurrentQuarter     int `json:"authorCountInCurrentQuarter"`
}
