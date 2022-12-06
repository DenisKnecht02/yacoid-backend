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

func AddAuthorsRequests(authorApi *fiber.Router, validate *validator.Validate) {

	(*authorApi).Post("/page_count", func(ctx *fiber.Ctx) error {

		request := new(types.AuthorPageCountRequest)

		if err := ctx.BodyParser(request); err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		validateErrors := request.Validate(validate)

		if validateErrors != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{
				Error: "Error on fields: " + strings.Join(validateErrors, ", "),
			})
		}

		count, err := database.GetAuthorPageCount(request)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Data: bson.M{
				"count": count,
			},
		})

	})

	(*authorApi).Post("/page", func(ctx *fiber.Ctx) error {

		request := new(types.AuthorPageRequest)

		if err := ctx.BodyParser(request); err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		validateErrors := request.Validate(validate)

		if validateErrors != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{
				Error: "Error on fields: " + strings.Join(validateErrors, ", "),
			})
		}

		authors, err := database.GetAuthors(request.PageSize, request.Page, request.Filter, request.Sort)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Data: bson.M{
				"authors": authors,
			},
		})

	})

	(*authorApi).Post("/", func(ctx *fiber.Ctx) error {

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

		userId, err := auth.AuthenticateAndGetId(ctx)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Authentication failed", Error: err.Error()})
		}

		authorId, err := database.CreateAuthor(request, userId)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully created author!",
			Data: bson.M{
				"authorId": authorId.Hex(),
			},
		})
	})

	(*authorApi).Delete("/", func(ctx *fiber.Ctx) error {

		authorId, err := GetRequiredStringQuery(ctx.Query("id"))

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Author ID required", Error: err.Error()})
		}

		_, err = auth.Authenticate(ctx, constants.RoleModerator, constants.RoleAdmin)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Authentication failed", Error: err.Error()})
		}

		usedSources, err := database.DeleteAuthor(authorId)

		if err != nil {

			if err == constants.ErrorAuthorDeletionBecauseInUse {

				return ctx.Status(GetErrorCode(err)).JSON(Response{
					Error: err.Error(),
					Data: bson.M{
						"sources": usedSources,
					},
				})

			} else {
				return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Deletion failed", Error: err.Error()})
			}

		}

		return ctx.JSON(Response{
			Message: "Successfully deleted author!",
		})
	})

	(*authorApi).Put("/", func(ctx *fiber.Ctx) error {

		request := new(types.ChangeAuthorRequest)

		if err := ctx.BodyParser(request); err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		_, err := types.ParseStringToAuthorType(request.Type.String())

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		validateErrors := request.Validate(validate)

		if validateErrors != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{
				Error: "Error on fields: " + strings.Join(validateErrors, ", "),
			})
		}

		userId, err := auth.AuthenticateAndGetId(ctx)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Authentication failed", Error: err.Error()})
		}

		err = database.ChangeAuthor(request, userId, validate)

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Message: "Change failed", Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully changed author!",
		})

	})

}
