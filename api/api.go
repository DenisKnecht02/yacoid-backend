package api

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"yacoid_server/auth"
	"yacoid_server/common"
	"yacoid_server/constants"
	"yacoid_server/database"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func StartAPI() {

	InitErrorCodeMap()

	fmt.Println("Starting server...")

	app := fiber.New()

	api := app.Group("/api")

	api.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	v1 := api.Group("/v1")

	validate := validator.New()

	definitionApi := v1.Group("/definitions")
	AddDefinitionRequests(&definitionApi, validate)

	authorApi := v1.Group("/authors")
	AddAuthorsRequests(&authorApi, validate)

	sourceApi := v1.Group("/sources")
	AddSourcesRequests(&sourceApi, validate)

	commonApi := v1.Group("/common")
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

func AuthMiddleware(roles ...constants.Role) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {

		_, err := auth.Authenticate(ctx, roles...)

		if err != nil {
			return err
		}

		return ctx.Next()

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
	ErrorCodeMap[common.ErrorAuthorNotFound] = fiber.StatusBadRequest
	ErrorCodeMap[common.ErrorNotEnoughPermissions] = fiber.StatusUnauthorized
	ErrorCodeMap[ErrorEmailVerification] = fiber.StatusBadRequest
	ErrorCodeMap[ErrorChangePassword] = fiber.StatusBadRequest

	ErrorCodeMap[database.ErrorDefinitionNotFound] = fiber.StatusNotFound
	ErrorCodeMap[database.ErrorDefinitionAlreadyApproved] = fiber.StatusBadRequest
	ErrorCodeMap[database.ErrorDefinitionRejectionBelongsToAnotherUser] = fiber.StatusUnauthorized
	ErrorCodeMap[database.ErrorDefinitionRejectionNotAnsweredYet] = fiber.StatusBadRequest
	ErrorCodeMap[fiber.ErrUnauthorized] = fiber.StatusUnauthorized

}
