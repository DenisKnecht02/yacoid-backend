package common

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var ErrorValidation = errors.New("INVALID_INPUT")
var ErrorInvalidType = errors.New("INVALID_TYPE")
var ErrorNotFound = errors.New("ENTITY_NOT_FOUND")
var ErrorInvalidEnum = errors.New("INVALID_ENUM")

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

func LoadEnvironmentVariables() error {

	err := godotenv.Load(".env")

	if err != nil {
		return err
	}

	return nil

}
