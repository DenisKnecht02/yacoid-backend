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

func AddSourcesRequests(api *fiber.Router, validate *validator.Validate) {

	(*api).Get("/source", func(ctx *fiber.Ctx) error {

		id := ctx.Query("id")

		source, err := database.GetSourceById(id)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		response, err := database.SourceToResponse(source)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Data: bson.M{"source": response},
		})

	})

	(*api).Post("/", func(ctx *fiber.Ctx) error {

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

	(*api).Delete("/", func(ctx *fiber.Ctx) error {

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

	(*api).Put("/", func(ctx *fiber.Ctx) error {

		request := new(types.ChangeSourceRequest)

		if err := ctx.BodyParser(request); err != nil {
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

		err = database.ChangeSource(request, id, validate)
		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully changed source!",
		})
	})

	(*api).Post("/page_count", func(ctx *fiber.Ctx) error {

		request := new(types.SourcePageCountRequest)

		if err := ctx.BodyParser(request); err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		validateErrors := request.Validate(validate)

		if validateErrors != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{
				Error: "Error on fields: " + strings.Join(validateErrors, ", "),
			})
		}

		if request.Filter == nil {
			request.Filter = &types.SourceFilter{}
		}
		
		if request.Filter.Approved != nil && *request.Filter.Approved == false {
			
			_, err := auth.AuthenticateAndGetId(ctx, constants.RoleModerator, constants.RoleAdmin)

			if err != nil {
				return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Authentication failed", Error: err.Error()})
			}
			
		}

		count, err := database.GetSourcePageCount(request)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Data: bson.M{
				"count": count,
			},
		})

	})

	(*api).Post("/page", func(ctx *fiber.Ctx) error {

		request := new(types.SourcePageRequest)

		if err := ctx.BodyParser(request); err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		validateErrors := request.Validate(validate)

		if validateErrors != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{
				Error: "Error on fields: " + strings.Join(validateErrors, ", "),
			})
		}

		if request.Filter == nil {
			request.Filter = &types.SourceFilter{}
		}
		
		if request.Filter.Approved != nil && *request.Filter.Approved == false {
			
			_, err := auth.AuthenticateAndGetId(ctx, constants.RoleModerator, constants.RoleAdmin)

			if err != nil {
				return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Authentication failed", Error: err.Error()})
			}
			
		}

		sources, err := database.GetSources(request)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		responses, err := database.SourcesToResponses(&sources)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Data: bson.M{
				"sources": responses,
			},
		})

	})

}
