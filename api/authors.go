package api

import (
	"strings"
	"yacoid_server/database"
	"yacoid_server/types"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func AddAuthorsRequests(authorApi *fiber.Router, validate *validator.Validate) {

	(*authorApi).Post("/create", func(ctx *fiber.Ctx) error {

		request := new(types.CreateAuthorRequest)

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
		err := database.CreateAuthor(request, authToken)
		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully created author!",
		})
	})

}
