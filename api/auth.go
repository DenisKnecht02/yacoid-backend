package api

import (
	"fmt"
	"strings"
	"yacoid_server/database"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func AddAuthRequests(authApi *fiber.Router, validate *validator.Validate) {

	(*authApi).Post("/register", func(ctx *fiber.Ctx) error {

		input := new(database.User)

		if err := ctx.BodyParser(input); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{Error: err.Error()})
		}

		validateErrors := input.Validate(validate)

		if validateErrors != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{
				Error: "Error on fields: " + strings.Join(validateErrors, ", "),
			})
		}

		user, err := database.Register(*input)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{Error: err.Error()})
		}
		return ctx.JSON(Response{Data: user})

	})

	(*authApi).Get("/login/:email/:password", func(ctx *fiber.Ctx) error {

		email := ctx.Params("email")
		password := ctx.Params("password")
		fmt.Println("Input", email, password)
		user, err := database.Login(email, password)
		fmt.Println(user, err)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{Error: err.Error()})
		}

		/* Hide some attributes */
		user.PasswordSalt = ""
		user.PasswordHash = ""

		return ctx.JSON(Response{
			Data: bson.M{"user: ": user},
		})

	})

	(*authApi).Get("/password_salt/:email", func(ctx *fiber.Ctx) error {

		email := ctx.Params("email")
		fmt.Println("Input", email)
		salt, err := database.GetPasswordSalt(email)
		fmt.Println("SALT", salt, err)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{Error: err.Error()})
		}
		return ctx.JSON(Response{
			Data: bson.M{"passwordSalt": salt},
		})

	})

	(*authApi).Get("/logout", func(ctx *fiber.Ctx) error {

		err := database.Logout(ctx.GetReqHeaders()["Authtoken"])

		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully logged out!",
		})

	})

	(*authApi).Get("/request_password_reset/:email", func(ctx *fiber.Ctx) error {

		email := ctx.Params("email")
		token, err := database.InitiatePasswordReset(email)

		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Data: bson.M{"token": token},
		})

	})

	(*authApi).Get("/reset_password/:token/:password_hash", func(ctx *fiber.Ctx) error {

		token := ctx.Params("token")
		passwordHash := ctx.Params("password_hash")
		err := database.ResetPassword(token, passwordHash)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully reset password!",
		})

	})

}
