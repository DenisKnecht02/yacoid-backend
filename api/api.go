package api

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"yacoid_server/auth"
	"yacoid_server/constants"
	"yacoid_server/types"

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
		AllowOrigins: strings.TrimSuffix(strings.Replace(os.Getenv(constants.EnvAuthRedirectUrl), "localhost", "127.0.0.1", 1), "/"),
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	v1 := api.Group("/v1")

	validate := validator.New()
	validate.RegisterValidation("is-author-type", ValidateAuthorType)
	validate.RegisterValidation("is-source-type", ValidateSourceType)
	validate.RegisterValidation("is-definition-category", ValidateDefinitionCategory)

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

func GetRequiredStringQuery(queryValue string) (string, error) {

	if len(queryValue) == 0 {
		return "", constants.ErrorQueryValueRequired
	}

	return queryValue, nil

}

func ValidateAuthorType(fieldLevel validator.FieldLevel) bool {

	_, err := types.ParseStringToAuthorType(fieldLevel.Field().String())
	return err == nil

}

func ValidateSourceType(fieldLevel validator.FieldLevel) bool {

	_, err := types.ParseStringToSourceType(fieldLevel.Field().String())
	return err == nil

}

func ValidateDefinitionCategory(fieldLevel validator.FieldLevel) bool {

	_, err := types.ParseStringToDefinitionCategory(fieldLevel.Field().String())
	return err == nil

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

var ErrorCodeMap map[error]int = map[error]int{}

func GetErrorCode(err error) int {

	fmt.Printf("[%T] %v\n", err, err)

	/* mongo.CommandError will cause an error if using as map key -> "panic: runtime error: hash of unhashable type mongo.CommandError" */
	if strings.HasPrefix(err.Error(), "(Location") || strings.HasPrefix(err.Error(), "(BadValue)") || strings.HasPrefix(err.Error(), "(IndexNotFound)") {
		// ^-- not "(Location)", because error prefix looks like this: "(Location: Error <number>)"
		return fiber.StatusInternalServerError
	}

	code, exists := ErrorCodeMap[err]

	if exists {
		fmt.Printf("Error \"%v\" does not have an error code assigned to it.\n", err)
		return code
	}

	return fiber.StatusInternalServerError

}

func InitErrorCodeMap() {

	ErrorCodeMap[constants.ErrorInvalidID] = fiber.StatusBadRequest
	ErrorCodeMap[constants.ErrorValidation] = fiber.StatusBadRequest
	ErrorCodeMap[constants.ErrorInvalidType] = fiber.StatusBadRequest
	ErrorCodeMap[constants.ErrorNotFound] = fiber.StatusBadRequest

	ErrorCodeMap[constants.ErrorUserNotFound] = fiber.StatusNotFound
	ErrorCodeMap[constants.ErrorAuthorNotFound] = fiber.StatusBadRequest
	ErrorCodeMap[constants.ErrorSourceNotFound] = fiber.StatusBadRequest
	ErrorCodeMap[constants.ErrorNotEnoughPermissions] = fiber.StatusUnauthorized
	ErrorCodeMap[constants.ErrorQueryValueRequired] = fiber.StatusBadRequest

	ErrorCodeMap[constants.ErrorDefinitionNotFound] = fiber.StatusNotFound
	ErrorCodeMap[constants.ErrorDefinitionAlreadyApproved] = fiber.StatusBadRequest
	ErrorCodeMap[constants.ErrorDefinitionRejectionBelongsToAnotherUser] = fiber.StatusUnauthorized
	ErrorCodeMap[constants.ErrorDefinitionRejectionNotAnsweredYet] = fiber.StatusBadRequest
	ErrorCodeMap[fiber.ErrUnauthorized] = fiber.StatusUnauthorized

}
