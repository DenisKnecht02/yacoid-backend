package api

import (
	"strings"
	"yacoid_server/auth"
	"yacoid_server/constants"
	"yacoid_server/database"
	"yacoid_server/types"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func AddSourcesRequests(sourceApi *fiber.Router, validate *validator.Validate) {

	(*sourceApi).Post("/", func(ctx *fiber.Ctx) error {

		request := new(types.CreateSourceRequest)

		if err := ctx.BodyParser(request); err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		_, err := types.ParseStringToSourceType(request.Type.String())

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		validateErrors := request.Validate(validate)

		if validateErrors != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{
				Error: "Error on fields: " + strings.Join(validateErrors, ", "),
			})
		}

		id, err := auth.AuthenticateAndGetId(ctx)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Authentication failed", Error: err.Error()})
		}

		sourceId, err := database.CreateSource(request, id)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully created source!",
			Data: bson.M{
				"sourceId": sourceId.Hex(),
			},
		})
	})

	(*sourceApi).Delete("/", func(ctx *fiber.Ctx) error {

		sourceId, err := GetRequiredStringQuery(ctx.Query("id"))

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Source ID required", Error: err.Error()})
		}

		_, err = auth.Authenticate(ctx, constants.RoleModerator, constants.RoleAdmin)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Authentication failed", Error: err.Error()})
		}

		usedDefinitions, err := database.DeleteSource(sourceId)

		if err != nil {

			if err == constants.ErrorSourceDeletionBecauseInUse {

				return ctx.Status(GetErrorCode(err)).JSON(Response{
					Error: err.Error(),
					Data: bson.M{
						"definitions": usedDefinitions,
					},
				})

			} else {
				return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Deletion failed", Error: err.Error()})
			}

		}

		return ctx.JSON(Response{
			Message: "Successfully deleted source!",
		})
	})

}
