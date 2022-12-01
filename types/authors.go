package types

import (
	"strings"
	"time"
	"yacoid_server/common"
	"yacoid_server/constants"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Author struct {
	ID                     primitive.ObjectID      `bson:"_id" json:"-"`
	SlugId                 string                  `bson:"slug_id" json:"slugId"`
	SubmittedBy            string                  `bson:"submitted_by" json:"submittedBy"`
	SubmittedDate          time.Time               `bson:"submitted_date" json:"submittedDate"`
	LastChangeDate         time.Time               `bson:"last_change_date" json:"lastChangeDate"`
	ApprovedBy             *string                 `bson:"approved_by" json:"approvedBy"`
	ApprovedDate           *time.Time              `bson:"approved_date" json:"approvedDate"`
	Approved               bool                    `bson:"approved" json:"approved"`
	Type                   AuthorType              `bson:"type" json:"type" validate:"required"`
	PersonProperties       *PersonProperties       `bson:"person_properties" json:"personProperties" validate:"required_without=OrganizationProperties,omitempty,dive"`
	OrganizationProperties *OrganizationProperties `bson:"organization_properties" json:"organizationProperties" validate:"required_without=PersonProperties,omitempty,dive"`
}

func (object *Author) Validate(validate *validator.Validate) []string {

	errorFields := common.ValidateStruct(object, validate)

	if object.Type == EnumAuthorType.Person && object.PersonProperties == nil {
		errorFields = append(errorFields, "PersonProperties missing")
	} else if object.Type == EnumAuthorType.Organization && object.OrganizationProperties == nil {
		errorFields = append(errorFields, "OrganizationProperties missing")
	}

	return errorFields
}

type PersonProperties struct {
	FirstName string `bson:"first_name" json:"firstName" validate:"required,min=1"`
	LastName  string `bson:"last_name" json:"lastName" validate:"required,min=1"`
}

type OrganizationProperties struct {
	OrganizationName string `bson:"organization_name" json:"organizationName" validate:"required,min=1"`
}

type CreateAuthorRequest struct {
	Type                   AuthorType              `bson:"type" json:"type" validate:"required"`
	PersonProperties       *PersonProperties       `bson:"person_properties" json:"personProperties" validate:"required_without=OrganizationProperties,omitempty,dive"`
	OrganizationProperties *OrganizationProperties `bson:"organization_properties" json:"organizationProperties" validate:"required_without=PersonProperties,omitempty,dive"`
}

func (object *CreateAuthorRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(object, validate)
}

type ChangeAuthorRequest struct {
	ID                     *string                       `json:"id" validate:"required"`
	Type                   *AuthorType                   `json:"type" validate:"omitempty"`
	PersonProperties       *ChangePersonProperties       `json:"personProperties" validate:"omitempty,dive"`
	OrganizationProperties *ChangeOrganizationProperties `json:"organizationProperties" validate:"omitempty,dive"`
}

func (object *ChangeAuthorRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(object, validate)
}

type ChangePersonProperties struct {
	FirstName *string `bson:"first_name" json:"firstName" validate:"omitempty,min=1"`
	LastName  *string `bson:"last_name" json:"lastName" validate:"omitempty,min=1"`
}

type ChangeOrganizationProperties struct {
	OrganizationName *string `bson:"organization_name" json:"organizationName" validate:"omitempty,min=1"`
}

type AuthorPageCountRequest struct {
	PageSize int           `json:"pageSize" validate:"required,min=1"`
	Filter   *AuthorFilter `json:"filter" validate:"omitempty,dive"`
}

func (request *AuthorPageCountRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(request, validate)
}

type AuthorPageRequest struct {
	PageSize int           `json:"pageSize" validate:"required,min=1"`
	Page     int           `json:"page" validate:"required,min=1"`
	Filter   *AuthorFilter `json:"filter" validate:"omitempty,dive"`
	Sort     *interface{}  `json:"sort"`
}

func (request *AuthorPageRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(request, validate)
}

type AuthorFilter struct {
	FirstName        *string `json:"firstName" validate:"omitempty,min=1"`
	LastName         *string `json:"lastName" validate:"omitempty,min=1"`
	OrganizationName *string `json:"organizationName" validate:"omitempty,min=1"`
	Type             *string `json:"type" validate:"omitempty,min=1"`
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
		return authorType, constants.ErrorInvalidEnum
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
