package api

import (
	"fmt"
	"yacoid_server/database"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

type DeleteUserRequest struct {
	PasswordHash string `bson:"password_hash,omitempty" json:"passwordHash,omitempty"`
	Reason       string `bson:"reason,omitempty" json:"reason,omitempty"`
}

type ChangeAccountDataRequest struct {
	FirstName       *string `bson:"first_name,omitempty" json:"firstName,omitempty"`
	LastName        *string `bson:"last_name,omitempty" json:"lastName,omitempty"`
	Email           *string `bson:"email,omitempty" json:"email,omitempty"`
	City            *string `bson:"city,omitempty" json:"city,omitempty"`
	CurrentPassword *string `bson:"current_password,omitempty" json:"currentPassword,omitempty"`
	NewPassword     *string `bson:"new_password,omitempty" json:"newPassword,omitempty"`
}

func AddUserRequests(userApi *fiber.Router, validate *validator.Validate) {

	(*userApi).Post("/delete_user", func(ctx *fiber.Ctx) error {

		request := new(DeleteUserRequest)

		if err := ctx.BodyParser(request); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{Error: err.Error()})
		}

		authToken := ctx.GetReqHeaders()["Authtoken"]
		err := database.DeleteUser(authToken, request.PasswordHash, request.Reason)

		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Message: "Successfully deleted user!",
		})

	})

	(*userApi).Post("/change_account_data", func(ctx *fiber.Ctx) error {

		request := new(ChangeAccountDataRequest)

		fmt.Println("1")
		if err := ctx.BodyParser(request); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{Error: err.Error()})
		}
		fmt.Println(request)

		authToken := ctx.GetReqHeaders()["Authtoken"]
		response, err := database.ChangeAccountData(authToken, request.FirstName, request.LastName, request.Email, request.City, request.CurrentPassword, request.NewPassword)

		if response.EmailVerification != nil && response.EmailVerification.Error != nil {
			errorText := ErrorEmailVerification.Error()
			response.EmailVerification.Error = &errorText
		}

		if response.ChangePassword != nil && response.ChangePassword.Error != nil {
			if *response.ChangePassword.Error != "INVALID_CREDENTIALS" {
				errorText := ErrorChangePassword.Error()
				response.ChangePassword.Error = &errorText
			}
		}

		fmt.Println(response, err)

		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(Response{Error: err.Error()})
		}

		return ctx.JSON(Response{
			Data: bson.M{
				"response": response,
			},
		})

	})
}
