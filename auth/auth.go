package auth

import (
	"fmt"
	"os"
	"strings"
	"yacoid_server/common"
	"yacoid_server/constants"

	"github.com/authorizerdev/authorizer-go"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/slices"
)

var AuthClient *authorizer.AuthorizerClient

func Initialize() error {

	defaultHeaders := map[string]string{}

	var err error
	AuthClient, err = authorizer.NewAuthorizerClient(os.Getenv(constants.EnvAuthClientId), os.Getenv(constants.EnvAuthUrl), os.Getenv(constants.EnvAuthRedirectUrl), defaultHeaders)

	if err != nil {
		return err
	}

	return nil

}

func GetUserByToken(token string) (*authorizer.User, error) {

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer: %s", token),
	}

	user, err := AuthClient.GetProfile(headers)

	if err != nil {
		return nil, err
	}

	return user, nil

}

func SplitAuthorizationHeader(header string) (string, bool) {

	split := strings.Split(header, " ")
	if len(split) == 2 && strings.ToLower(split[0]) == "bearer" && len(split[1]) > 0 {
		return split[1], true
	}

	return "", false

}

// Can be the id or access token
func GetAuthorizationToken(ctx *fiber.Ctx) (string, error) {

	authHeader := ctx.GetReqHeaders()["Authorization"]

	token, ok := SplitAuthorizationHeader(authHeader)
	if !ok {
		return "", fiber.ErrUnauthorized
	}

	return token, nil

}

func GetUserByContext(ctx *fiber.Ctx) (*authorizer.User, error) {

	token, err := GetAuthorizationToken(ctx)

	if err != nil {
		return nil, err
	}

	return GetUserByToken(token)

}

func Authenticate(ctx *fiber.Ctx, requiredRoles ...constants.Role) (map[string]interface{}, error) {

	token, err := GetAuthorizationToken(ctx)

	if err != nil {
		return nil, err
	}

	response, err := AuthClient.ValidateJWTToken(&authorizer.ValidateJWTTokenInput{
		TokenType: authorizer.TokenTypeIDToken,
		Token:     token,
	})

	if err != nil {
		return nil, err
	}

	if !response.IsValid {
		return nil, common.ErrorValidation
	}

	roleInterfaceArray, ok := response.Claims["role"].([]interface{})

	if !ok {
		return nil, common.ErrorRoleClaimCast
	}

	roles, err := common.InterfaceArrayToStringArray(roleInterfaceArray)

	if err != nil {
		return nil, err
	}

	if len(requiredRoles) == 0 {
		requiredRoles = []constants.Role{constants.RoleUser}
	}

	hasEnoughPermissions := false

	if len(requiredRoles) == 0 {
		hasEnoughPermissions = true // no roles required
	} else {
		for _, requiredRole := range requiredRoles {
			if slices.Contains(roles, requiredRole.String()) {
				hasEnoughPermissions = true
				break
			}
		}
	}

	if !hasEnoughPermissions {
		return nil, common.ErrorNotEnoughPermissions
	}

	return response.Claims, nil

}

func AuthenticateAndGetId(ctx *fiber.Ctx, roles ...constants.Role) (string, error) {

	claims, err := Authenticate(ctx, roles...)

	if err != nil {
		return "", err
	}

	id, ok := claims["id"].(string)

	if !ok {
		return "", common.ErrorUserIdCast
	}

	return id, nil

}
