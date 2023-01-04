package types

import (
	"strings"
	"time"
	"yacoid_server/common"
	"yacoid_server/constants"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DefinitionsOfUserResponse struct {
	ID            primitive.ObjectID    `bson:"_id" json:"id"`
	SubmittedBy   string                `bson:"submitted_by" json:"submittedBy"` // user name instead of id
	SubmittedDate time.Time             `bson:"submitted_date" json:"submittedDate"`
	ApprovedBy    *string               `bson:"approved_by" json:"approvedBy"`
	ApprovedDate  *time.Time            `bson:"approved_date" json:"approvedDate"`
	Approved      bool                  `bson:"approved" json:"approved"`
	RejectionLog  *[]*RejectionResponse `bson:"rejection_log" json:"rejectionLog"`
	Content       string                `bson:"content" json:"content"`
	Source        SourceResponse        `bson:"source" json:"source"`
	Category      DefinitionCategory    `bson:"category" json:"category"`
	Status        DefinitionStatus      `bson:"status" json:"status"`
}

type DefinitionResponse struct {
	ID            primitive.ObjectID `bson:"_id" json:"id"`
	SubmittedBy   string             `bson:"submitted_by" json:"submittedBy"` // user name instead of id
	SubmittedDate time.Time          `bson:"submitted_date" json:"submittedDate"`
	Content       string             `bson:"content" json:"content"`
	Source        SourceResponse     `bson:"source" json:"source"`
	Category      DefinitionCategory `bson:"category" json:"category"`
}

type Definition struct {
	ID             primitive.ObjectID `bson:"_id" json:"id"`
	SubmittedBy    string             `bson:"submitted_by" json:"submittedBy"`
	SubmittedDate  time.Time          `bson:"submitted_date" json:"submittedDate"`
	LastChangeDate time.Time          `bson:"last_change_date" json:"lastChangeDate"`
	ApprovedBy     *string            `bson:"approved_by" json:"approvedBy"`
	ApprovedDate   *time.Time         `bson:"approved_date" json:"approvedDate"`
	Approved       bool               `bson:"approved" json:"approved"`
	RejectionLog   *[]*Rejection      `bson:"rejection_log" json:"-"`
	Content        string             `bson:"content" json:"content"`
	Source         primitive.ObjectID `bson:"source" json:"source"`
	Category       DefinitionCategory `bson:"category" json:"category"`
}

func (definition *Definition) IsApproved() bool {
	return definition.ApprovedBy != nil && definition.ApprovedDate != nil
}

func (definition *Definition) GetStatus() DefinitionStatus {

	if definition.Approved {
		return EnumDefinitionStatus.Approved
	}

	if len(*definition.RejectionLog) == 0 {
		return EnumDefinitionStatus.Pending
	}

	latestRejectionDate := definition.GetLatestRejection()

	// user submitted new change, that was not approved/rejected yet
	if !latestRejectionDate.IsZero() && latestRejectionDate.Before(definition.LastChangeDate) {
		return EnumDefinitionStatus.Pending
	}

	return EnumDefinitionStatus.Declined

}

func (definition *Definition) GetLatestRejection() time.Time {

	var latestRejectionDate time.Time
	for _, d := range *definition.RejectionLog {
		if d.RejectedDate.After(latestRejectionDate) {
			latestRejectionDate = d.RejectedDate
		}
	}

	return latestRejectionDate

}

type RejectionResponse struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	RejectedBy   string             `bson:"rejected_by" json:"rejectedBy" validate:"required"` // user name instead of id
	RejectedDate time.Time          `bson:"rejected_date" json:"rejectedDate" validate:"required"`
	Content      string             `bson:"content" json:"content" validate:"required"`
}

type Rejection struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	RejectedBy   string             `bson:"rejected_by" json:"rejectedBy" validate:"required"`
	RejectedDate time.Time          `bson:"rejected_date" json:"rejectedDate" validate:"required"`
	Content      string             `bson:"content" json:"content" validate:"required"`
}

type SubmitDefinitionRequest struct {
	Content  string             `json:"content" validate:"required,min=1"`
	SourceId string             `json:"sourceId" validate:"required"`
	Category DefinitionCategory `json:"category" validate:"required,is-definition-category"`
}

func (request *SubmitDefinitionRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(request, validate)
}

type DefinitionPageCountRequest struct {
	PageSize int               `json:"pageSize" validate:"required,min=1"`
	Filter   *DefinitionFilter `json:"filter" validate:"omitempty,dive"`
}

func (request *DefinitionPageCountRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(request, validate)
}

type DefinitionPageRequest struct {
	PageSize int               `json:"pageSize" validate:"required,min=1"`
	Page     int               `json:"page" validate:"required,min=1"`
	Filter   *DefinitionFilter `json:"filter" validate:"omitempty,dive"`
}

func (request *DefinitionPageRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(request, validate)
}

type RejectRequest struct {
	ID      string `json:"id" validate:"required"`
	Content string `json:"content" validate:"required,min=1"`
}

func (request *RejectRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(request, validate)
}

type ChangeDefinitionRequest struct {
	ID       *string             `json:"id" validate:"required"`
	Content  *string             `json:"content" validate:"omitempty,min=1"`
	SourceId *string             `json:"sourceId" validate:"omitempty,min=1"`
	Category *DefinitionCategory `json:"category" validate:"omitempty,is-definition-category"`
}

func (request *ChangeDefinitionRequest) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(request, validate)
}

type DefinitionFilter struct {
	Approved        *bool                 `json:"approved" bson:"approved" validate:"omitempty"`
	Content         *string               `json:"content" bson:"content" validate:"omitempty,min=1"`
	Categories      *[]DefinitionCategory `json:"categories" bson:"categories" validate:"omitempty,dive,is-definition-category"`
	AuthorIds       *[]string             `json:"authors" bson:"authors" validate:"omitempty,min=1"`
	PublishingYears *[]int                `json:"publishingYears" bson:"publishing_years" validate:"omitempty,min=1"`
	UserId          *string               `json:"userId" bson:"user_id" validate:"omitempty,min=1"`
}

type DefinitionCategory string

type definitionCategoryList struct {
	Unknown                DefinitionCategory
	HumanIntelligence      DefinitionCategory
	ArtificialIntelligence DefinitionCategory
	MachineIntelligence    DefinitionCategory
	PlantIntelligence      DefinitionCategory
	AlienIntelligence      DefinitionCategory
}

var EnumDefinitionCategory = &definitionCategoryList{
	Unknown:                "unknown",
	HumanIntelligence:      "human_intelligence",
	ArtificialIntelligence: "artificial_intelligence",
	MachineIntelligence:    "machine_intelligence",
	PlantIntelligence:      "plant_intelligence",
	AlienIntelligence:      "alien_intelligence",
}

var definitionCategoryMap = map[string]DefinitionCategory{
	"human_intelligence":      EnumDefinitionCategory.HumanIntelligence,
	"artificial_intelligence": EnumDefinitionCategory.ArtificialIntelligence,
	"machine_intelligence":    EnumDefinitionCategory.MachineIntelligence,
	"plant_intelligence":      EnumDefinitionCategory.PlantIntelligence,
	"alien_intelligence":      EnumDefinitionCategory.AlienIntelligence,
}

func ParseStringToDefinitionCategory(str string) (DefinitionCategory, error) {
	definitionCategory, ok := definitionCategoryMap[strings.ToLower(str)]
	if ok {
		return definitionCategory, nil
	} else {
		return definitionCategory, constants.ErrorInvalidEnum
	}
}

func (definitionCategory DefinitionCategory) String() string {
	switch definitionCategory {
	case EnumDefinitionCategory.HumanIntelligence:
		return "human_intelligence"
	case EnumDefinitionCategory.ArtificialIntelligence:
		return "artificial_intelligence"
	case EnumDefinitionCategory.MachineIntelligence:
		return "machine_intelligence"
	case EnumDefinitionCategory.PlantIntelligence:
		return "plant_intelligence"
	case EnumDefinitionCategory.AlienIntelligence:
		return "alien_intelligence"

	}
	return "unknown"
}

type DefinitionStatus string

type definitionStatusList struct {
	Unknown  DefinitionStatus
	Approved DefinitionStatus
	Pending  DefinitionStatus
	Declined DefinitionStatus
}

var EnumDefinitionStatus = &definitionStatusList{
	Unknown:  "unknown",
	Approved: "approved",
	Pending:  "pending",
	Declined: "declined",
}

var definitionStatusMap = map[string]DefinitionStatus{
	"approved": EnumDefinitionStatus.Approved,
	"pending":  EnumDefinitionStatus.Pending,
	"declined": EnumDefinitionStatus.Declined,
}

func ParseStringToDefinitionStatus(str string) (DefinitionStatus, error) {
	definitionStatus, ok := definitionStatusMap[strings.ToLower(str)]
	if ok {
		return definitionStatus, nil
	} else {
		return definitionStatus, constants.ErrorInvalidEnum
	}
}

func (definitionStatus DefinitionStatus) String() string {
	switch definitionStatus {
	case EnumDefinitionStatus.Approved:
		return "approved"
	case EnumDefinitionStatus.Pending:
		return "pending"
	case EnumDefinitionStatus.Declined:
		return "declined"
	}
	return "unknown"
}
