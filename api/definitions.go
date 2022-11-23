package api

import (
	"fmt"
	"strings"
	"yacoid_server/auth"
	"yacoid_server/constants"
	"yacoid_server/database"
	"yacoid_server/types"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func AddDefinitionRequests(definitionApi *fiber.Router, validate *validator.Validate) {

	(*definitionApi).Get("/definition/:id", func(ctx *fiber.Ctx) error {

		id := ctx.Params("id")

		definition, err := database.GetDefinitionById(id)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Data: bson.M{"definiton": definition},
		})

	})

	(*definitionApi).Post("/submit", AuthMiddleware(), func(ctx *fiber.Ctx) error {

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

		idToken, err := auth.GetIdTokenAndExpectRoleFromContext(ctx, constants.RoleUser)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		if err := ctx.BodyParser(request); err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		definition, err := database.SubmitDefinition(request, idToken)
		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Data: bson.M{"definiton": definition},
		})
	})

	(*definitionApi).Get("/approve/:id", func(ctx *fiber.Ctx) error {

		definitionId := ctx.Params("id")

		authToken := ctx.GetReqHeaders()["Authtoken"]
		err := database.ApproveDefinition(definitionId, authToken)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully approved definition!",
		})

	})

	(*definitionApi).Post("/reject", func(ctx *fiber.Ctx) error {

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

		authToken := ctx.GetReqHeaders()["Authtoken"]
		err := database.RejectDefinition(request.ID, authToken, request.Content)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully rejected definition!",
		})
	})

	(*definitionApi).Post("/change", func(ctx *fiber.Ctx) error {

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

		authToken := ctx.GetReqHeaders()["Authtoken"]
		err := database.ChangeDefinition(request.ID, request.Title, request.Content, request.Source, request.Tags, authToken)
		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully changed definition!",
		})
	})

	(*definitionApi).Get("/newest_definitions", func(ctx *fiber.Ctx) error {

		limit := GetOptionalIntParam(ctx.Query("limit"), 4)

		definitions, err := database.GetNewestDefinitions(limit)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Data: bson.M{
				"definitions": definitions,
			},
		})

	})

	(*definitionApi).Get("/page_count", func(ctx *fiber.Ctx) error {

		pageSize := GetOptionalIntParam(ctx.Query("page_size"), 4)
		count, err := database.GetPageCount(pageSize, bson.M{})

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Data: bson.M{
				"count": count,
			},
		})

	})

	(*definitionApi).Post("/page", func(ctx *fiber.Ctx) error {

		request := new(types.DefinitionPageRequest)

		if err := ctx.BodyParser(request); err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		fmt.Println(request)
		validateErrors := request.Validate(validate)

		if validateErrors != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{
				Error: "Error on fields: " + strings.Join(validateErrors, ", "),
			})
		}

		definitions, err := database.GetDefinitions(request.PageSize, request.Page, request.Filter, request.Sort)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Data: bson.M{
				"definitions": definitions,
			},
		})

	})

}
