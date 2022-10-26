package api

import (
	"fmt"
	"strings"
	"yacoid_server/database"
	"yacoid_server/types"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func AddSourcesRequests(sourceApi *fiber.Router, validate *validator.Validate) {

	(*sourceApi).Post("/create", func(ctx *fiber.Ctx) error {

		rawSourceType := ctx.Query("type")
		sourceType, err := types.ParseStringToSourceType(rawSourceType)
		fmt.Println("Source Type = " + sourceType.String())

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		request := new(types.CreateSourceRequest)

		if err := ctx.BodyParser(request); err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		validateErrors := request.Validate(validate)

		if validateErrors != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{
				Error: "Error on fields: " + strings.Join(validateErrors, ", "),
			})
		}

		authToken := ctx.GetReqHeaders()["Authtoken"]
		err = database.CreateSource(request, authToken)
		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully created source!",
		})
	})

}
