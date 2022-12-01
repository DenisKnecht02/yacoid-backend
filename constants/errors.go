package constants

import (
	"errors"
	"strings"
)

var ErrorInterfaceArrayToStringArrayCast = errors.New("FAILED_INTERFACE_ARRAY_TO_STRING_ARRAY_CAST")

var ErrorUserIdCast = errors.New("FAILED_USER_ID_CAST")
var ErrorRoleClaimCast = errors.New("FAILED_ROLE_CLAIM_CAST")
var ErrorNotEnoughPermissions = errors.New("NOT_ENOUGH_PERMISSIONS")
var ErrorUnexpectedSigningMethod = errors.New("UNEXPECTED_SIGNING_METHOD")
var ErrorMissingRole = errors.New("MISSING_ROLE")

var ErrorInvalidID = errors.New("INVALID_ID")
var ErrorValidation = errors.New("FAILED_VALIDATION")

func CreateValidationError(errorFields *[]string) error {
	return errors.New("VALIDATION_ERROR: Error on fields: " + strings.Join(*errorFields, ", "))
}

var ErrorInvalidType = errors.New("INVALID_TYPE")
var ErrorInvalidEnum = errors.New("INVALID_ENUM")
var ErrorInvalidValidationResponse = errors.New("INVALID_VALIDATION_RESPONSE")

var ErrorQueryValueRequired = errors.New("QUERY_VALUE_REQUIRED")

var ErrorAuthorCreation = errors.New("AUTHOR_CREATION")
var ErrorSourceCreation = errors.New("SOURCE_CREATION")

var ErrorAuthorAlreadyApproved = errors.New("AUTHOR_ALREADY_APPROVED")
var ErrorSourceAlreadyApproved = errors.New("SOURCE_ALREADY_APPROVED")
var ErrorDefinitionAlreadyApproved = errors.New("DEFINITION_ALREADY_APPROVED")
var ErrorDefinitionRejectionNotAnsweredYet = errors.New("DEFINITION_REJECTION_NOT_ANSWERED_YET")
var ErrorDefinitionRejectionBelongsToAnotherUser = errors.New("DEFINITION_REJECTION_BELONGS_TO_ANOTHER_USER")

var ErrorNotFound = errors.New("ENTITY_NOT_FOUND")
var ErrorUserNotFound = errors.New("USER_NOT_FOUND")
var ErrorSourceNotFound = errors.New("SOURCE_NOT_FOUND")
var ErrorAuthorNotFound = errors.New("AUTHOR_NOT_FOUND")
var ErrorDefinitionNotFound = errors.New("DEFINITION_NOT_FOUND")

var ErrorAuthorDeletionBecauseInUse = errors.New("AUTHOR_COULD_NOT_BE_DELETED_BECAUSE_IN_USE")
var ErrorAuthorChangeBecauseInUse = errors.New("AUTHOR_COULD_NOT_BE_CHANGED_BECAUSE_IN_USE")
var ErrorSourceDeletionBecauseInUse = errors.New("SOURCE_COULD_NOT_BE_DELETED_BECAUSE_IN_USE")
