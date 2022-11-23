package common

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var ErrorValidation = errors.New("INVALID_INPUT")
var ErrorInvalidType = errors.New("INVALID_TYPE")
var ErrorNotFound = errors.New("ENTITY_NOT_FOUND")
var ErrorInvalidEnum = errors.New("INVALID_ENUM")
var ErrorUnexpectedSigningMethod = errors.New("UNEXPECTED_SIGNING_METHOD")
var ErrorMissingRole = errors.New("MISSING_ROLE")
var ErrorValidationResponseInvalid = errors.New("VALIDATION_RESPONSE_INVALID")

func ValidateStruct(s interface{}, validate *validator.Validate) []string {

	errorFields := []string{}
	err := validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errorMessage := err.StructNamespace() + " (tag:" + err.Tag() + ", should_be:" + err.Param() + ")"
			errorFields = append(errorFields, errorMessage)
		}
	}

	if len(errorFields) == 0 {
		return nil
	}
	return errorFields

}

func InterfaceArrayToStringArray(dataArray []interface{}) ([]string, bool) {

	stringArray := []string{}

	for _, data := range dataArray {

		str, ok := data.(string)

		if !ok {
			return stringArray, false
		}

		stringArray = append(stringArray, str)

	}

	return stringArray, true

}

func GetCurrentQuarterDate() time.Time {

	now := time.Now()

	var quarterTime time.Time

	if now.Month() <= 3 {
		quarterTime = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
	} else if now.Month() <= 6 {
		quarterTime = time.Date(now.Year(), time.April, 1, 0, 0, 0, 0, now.Location())
	} else if now.Month() <= 6 {
		quarterTime = time.Date(now.Year(), time.July, 1, 0, 0, 0, 0, now.Location())
	} else {
		quarterTime = time.Date(now.Year(), time.October, 1, 0, 0, 0, 0, now.Location())
	}

	return quarterTime

}

func LoadEnvironmentVariables() error {

	err := godotenv.Load(".env")

	if err != nil {
		return err
	}

	return nil

}
