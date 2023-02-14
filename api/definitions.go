package api

import (
	"strings"
	"yacoid_server/auth"
	"yacoid_server/common"
	"yacoid_server/constants"
	"yacoid_server/database"
	"yacoid_server/types"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func AddDefinitionRequests(api *fiber.Router, validate *validator.Validate) {

	(*api).Get("/definition", func(ctx *fiber.Ctx) error {

		id := ctx.Query("id")

		definition, err := database.GetDefinitionById(id)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		response, err := database.DefinitionToResponse(definition)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Data: bson.M{"definition": response},
		})

	})

	(*api).Post("/submit", func(ctx *fiber.Ctx) error {

		request := new(types.SubmitDefinitionRequest)

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

		definition, err := database.SubmitDefinition(request, id)
		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully created definition!",
			Data: bson.M{
				"definitionId": definition.Hex(),
			},
		})
	})

	(*api).Get("/approve", func(ctx *fiber.Ctx) error {

		definitionId := ctx.Query("id")

		id, err := auth.AuthenticateAndGetId(ctx, constants.EnumRole.Moderator, constants.EnumRole.Admin)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Authentication failed", Error: err.Error()})
		}

		err = database.ApproveDefinition(definitionId, id)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully approved definition!",
		})

	})

	(*api).Post("/reject", func(ctx *fiber.Ctx) error {

		request := new(types.RejectRequest)

		if err := ctx.BodyParser(request); err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		validateErrors := request.Validate(validate)

		if validateErrors != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{
				Error: "Error on fields: " + strings.Join(validateErrors, ", "),
			})
		}

		id, err := auth.AuthenticateAndGetId(ctx, constants.EnumRole.Moderator, constants.EnumRole.Admin)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Authentication failed", Error: err.Error()})
		}

		err = database.RejectDefinition(request.ID, request.Content, id)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully rejected definition!",
		})
	})

	(*api).Put("/", func(ctx *fiber.Ctx) error {

		request := new(types.ChangeDefinitionRequest)

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

		err = database.ChangeDefinition(request, id)
		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully changed definition!",
		})
	})

	(*api).Get("/newest_definitions", func(ctx *fiber.Ctx) error {

		limit := GetOptionalIntParam(ctx.Query("limit"), 4)

		definitions, err := database.GetNewestDefinitions(limit)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		responses, err := database.DefinitionsToResponses(&definitions)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Data: bson.M{
				"definitions": responses,
			},
		})

	})

	(*api).Post("/page_count", func(ctx *fiber.Ctx) error {

		request := new(types.DefinitionPageCountRequest)

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
			request.Filter = &types.DefinitionFilter{}
		}

		// user wants to see (only) submitted definitions
		if request.Filter.Approved == nil || *request.Filter.Approved == false {

			id, roles, err := auth.Authenticate(ctx)

			if err != nil {
				return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Authentication failed", Error: err.Error()})
			}

			// if the user is not an moderator or admin, he can only view unapproved definitions of himself
			if !common.ArrayContainsOr(roles, constants.EnumRole.Moderator, constants.EnumRole.Admin) {
				if request.Filter.UserId == nil || *request.Filter.UserId != id {
					return ctx.Status(GetErrorCode(constants.ErrorNotEnoughPermissions)).JSON(Response{Message: "Authentication failed", Error: constants.ErrorNotEnoughPermissions.Error()})
				}
			}

		}

		count, err := database.GetDefinitionPageCount(request)

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

		request := new(types.DefinitionPageRequest)

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
			request.Filter = &types.DefinitionFilter{}
		}

		// user wants to see (only) submitted definitions
		if request.Filter.Approved == nil || *request.Filter.Approved == false {

			id, roles, err := auth.Authenticate(ctx)

			if err != nil {
				return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Authentication failed", Error: err.Error()})
			}

			// if the user is not an moderator or admin, he can only view unapproved definitions of himself
			if !common.ArrayContainsOr(roles, constants.EnumRole.Moderator, constants.EnumRole.Admin) {
				if request.Filter.UserId == nil || *request.Filter.UserId != id {
					return ctx.Status(GetErrorCode(constants.ErrorNotEnoughPermissions)).JSON(Response{Message: "Authentication failed", Error: constants.ErrorNotEnoughPermissions.Error()})
				}
			}

		}

		definitions, err := database.GetDefinitions(request)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		var responses interface{}

		// if the user wants to see his own definitions, they will have more information in it
		if request.Filter.UserId != nil {

			id, roles, err := auth.Authenticate(ctx)

			if err == nil && (id == *request.Filter.UserId || common.ArrayContainsOr(roles, constants.EnumRole.Moderator, constants.EnumRole.Admin)) {
				responses, err = database.DefinitionsToUserResponses(&definitions)
			}

		} else if request.AdminInformation != nil && *request.AdminInformation == true {

			_, _, err := auth.Authenticate(ctx, constants.EnumRole.Moderator, constants.EnumRole.Admin)

			if err == nil {
				responses, err = database.DefinitionsToUserResponses(&definitions)
			}

		}

		if responses == nil {

			responses, err = database.DefinitionsToResponses(&definitions)

			if err != nil {
				return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
			}

		}

		return ctx.JSON(Response{
			Data: bson.M{
				"definitions": responses,
			},
		})

	})

	(*api).Delete("/", func(ctx *fiber.Ctx) error {

		definitionId, err := GetRequiredStringQuery(ctx.Query("id"))

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Definition ID required", Error: err.Error()})
		}

		_, _, err = auth.Authenticate(ctx, constants.EnumRole.Moderator, constants.EnumRole.Admin)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Authentication failed", Error: err.Error()})
		}

		err = database.DeleteDefinition(definitionId)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Deletion failed", Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully deleted definition!",
		})
	})

}
