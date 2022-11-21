package api

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"yacoid_server/common"
	"yacoid_server/constants"
	"yacoid_server/database"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func StartAPI() {

	InitErrorCodeMap()

	fmt.Println("Starting server...")

	app := fiber.New()

	api := app.Group("/api")

	validate := validator.New()

	definitionApi := api.Group("/definitions")
	AddDefinitionRequests(&definitionApi, validate)

	authorApi := api.Group("/authors")
	AddAuthorsRequests(&authorApi, validate)

	sourceApi := api.Group("/sources")
	AddSourcesRequests(&sourceApi, validate)

	authApi := api.Group("/auth")
	AddAuthRequests(&authApi, validate)

	userApi := api.Group("/user")
	AddAuthRequests(&userApi, validate)

	commonApi := api.Group("/common")
	AddCommonRequests(&commonApi, validate)

	fmt.Println("Started server on port " + os.Getenv(constants.EnvKeyRestPort))

	app.Listen(":" + os.Getenv(constants.EnvKeyRestPort))

}

func GetOptionalIntParam(stringValue string, defaultValue int) int {

	if len(stringValue) == 0 {
		return defaultValue
	} else {
		tempLimit, err := strconv.Atoi(stringValue)

		if err != nil {
			return defaultValue
		}

		return tempLimit

	}

}

type Response struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

var ErrorEmailVerification = errors.New("EMAIL_VERIFICATION_ERROR")
var ErrorChangePassword = errors.New("CHANGE_PASSWORD_ERROR")

var ErrorCodeMap map[error]int = map[error]int{}

func GetErrorCode(err error) int {

	code, exists := ErrorCodeMap[err]

	if exists {
		return code
	}

	return fiber.StatusInternalServerError

}

func InitErrorCodeMap() {

	ErrorCodeMap[database.InvalidID] = fiber.StatusBadRequest
	ErrorCodeMap[common.ErrorValidation] = fiber.StatusBadRequest
	ErrorCodeMap[common.ErrorInvalidType] = fiber.StatusBadRequest
	ErrorCodeMap[common.ErrorNotFound] = fiber.StatusBadRequest

	ErrorCodeMap[database.ErrorUserNotFound] = fiber.StatusNotFound
	ErrorCodeMap[database.ErrorNotEnoughPermissions] = fiber.StatusUnauthorized
	ErrorCodeMap[database.ErrorInvalidCredentials] = fiber.StatusUnauthorized
	ErrorCodeMap[database.ErrorPasswordResetExpiryDateExceeded] = fiber.StatusBadRequest
	ErrorCodeMap[database.ErrorUserAlreadyExists] = fiber.StatusBadRequest
	ErrorCodeMap[database.ErrorUserAlreadyLoggedIn] = fiber.StatusBadRequest
	ErrorCodeMap[ErrorEmailVerification] = fiber.StatusBadRequest
	ErrorCodeMap[ErrorChangePassword] = fiber.StatusBadRequest

	ErrorCodeMap[database.ErrorDefinitionNotFound] = fiber.StatusNotFound
	ErrorCodeMap[database.ErrorDefinitionAlreadyApproved] = fiber.StatusBadRequest
	ErrorCodeMap[database.ErrorDefinitionRejectionBelongsToAnotherUser] = fiber.StatusUnauthorized
	ErrorCodeMap[database.ErrorDefinitionRejectionNotAnsweredYet] = fiber.StatusBadRequest

}
