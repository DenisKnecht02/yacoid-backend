package api

import (
	"yacoid_server/database"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func AddCommonRequests(api *fiber.Router, validate *validator.Validate) {

	(*api).Get("/statistics", func(ctx *fiber.Ctx) error {

		response, err := database.GetStatistics()

		if err != nil {
			return ctx.Status(GetErrorCode(err)).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Data: response,
		})
	})

}
