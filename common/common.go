package common

import (
	"time"
	"yacoid_server/constants"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

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

func InterfaceArrayToStringArray(dataArray []interface{}) ([]string, error) {

	stringArray := []string{}

	for _, data := range dataArray {

		str, ok := data.(string)

		if !ok {
			return stringArray, constants.ErrorInterfaceArrayToStringArrayCast
		}

		stringArray = append(stringArray, str)

	}

	return stringArray, nil

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
