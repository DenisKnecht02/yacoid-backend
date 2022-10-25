package types

import (
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

type CreateSourceRequest struct {
	Authors []string `bson:"authors" json:"authors" validate:"required,min=1"`
}

func (rejection *CreateSourceRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(rejection, validate)
}

type Author struct {
	ID            primitive.ObjectID `bson:"_id" json:"-"`
	SlugId        string             `bson:"slug_id" json:"slugId"`
	SubmittedBy   primitive.ObjectID `bson:"submitted_by" json:"submittedBy"`
	SubmittedDate time.Time          `bson:"submitted_date" json:"submittedDate"`
	FirstName     string             `bson:"first_name" json:"firstName"`
	LastName      string             `bson:"last_name" json:"lastName"`
}

type Source struct {
	ID            primitive.ObjectID   `bson:"_id" json:"-"`
	SubmittedBy   primitive.ObjectID   `bson:"submitted_by" json:"submittedBy"`
	SubmittedDate time.Time            `bson:"submitted_date" json:"submittedDate"`
	Authors       []primitive.ObjectID `bson:"authors" json:"authors" validate:"required,min=1"`
}

func (author *Source) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(author, validate)
}

type DefinitionFilter struct {
	Title           *string      `json:"title" bson:"title" validate:"omitempty"`
	Content         *string      `json:"content" bson:"content" validate:"omitempty"`
	PublishingDates *[]time.Time `json:"publishing_dates" bson:"publishing_dates" validate:"omitempty,min=1"`
	Authors         *[]*Author   `json:"authors" bson:"authors" validate:"omitempty,min=1,dive"`
	Sources         *[]*Source   `json:"sources" bson:"sources" validate:"omitempty,min=1,dive"`
	Tags            *[]string    `json:"tags" bson:"tags" validate:"omitempty,min=1"`
}
