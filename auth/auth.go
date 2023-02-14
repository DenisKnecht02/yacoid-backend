package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

func GetNicknameOfUser(userId string) (string, error) {

	user, err := GetUser(userId)

	if err != nil {
		return "", err
	}

	if user.Nickname == nil {
		return "anonymous", nil
	}

	return *user.Nickname, nil
}

func GetUser(userId string) (*authorizer.User, error) {

	query := fmt.Sprintf(
		`query {_user (params: {
				id: "%s"
			}){
				id
				email
				preferred_username
				email_verified
				signup_methods
				given_name
				family_name
				middle_name
				nickname
				picture
				gender
				birthdate
				phone_number
				phone_number_verified
				roles
				created_at
				updated_at
				is_multi_factor_auth_enabled
			}
		}`,
		userId)

	reqBody := map[string]string{
		"query": query,
	}

	jsonReq, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, os.Getenv(constants.EnvAuthUrl)+"/graphql", bytes.NewReader(jsonReq))

	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-authorizer-admin-secret", os.Getenv(constants.EnvAuthAdminSecret))

	res, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	resBody := map[string]interface{}{}
	json.Unmarshal(bodyBytes, &resBody)
	data, ok := resBody["data"].(map[string]interface{})

	if !ok {
		return nil, constants.ErrorUserNotFound
	}

	userData, ok := data["_user"].(map[string]interface{})

	if !ok {
		return nil, constants.ErrorUserNotFound
	}

	userBytes, err := json.Marshal(userData)

	if err != nil {
		return nil, err
	}

	var user authorizer.User
	json.Unmarshal(userBytes, &user)

	return &user, nil
}

func authenticate(ctx *fiber.Ctx, requiredRoles ...constants.Role) (map[string]interface{}, *[]constants.Role, error) {

	token, err := GetAuthorizationToken(ctx)

	if err != nil {
		return nil, nil, err
	}

	response, err := AuthClient.ValidateJWTToken(&authorizer.ValidateJWTTokenInput{
		TokenType: authorizer.TokenTypeIDToken,
		Token:     token,
	})

	if err != nil {
		return nil, nil, err
	}

	if !response.IsValid {
		return nil, nil, constants.ErrorValidation
	}

	roleInterfaceArray, ok := response.Claims["role"].([]interface{})

	if !ok {
		return nil, nil, constants.ErrorRoleClaimCast
	}

	rolesAsString, err := common.InterfaceArrayToStringArray(roleInterfaceArray)

	if err != nil {
		return nil, nil, err
	}

	roles, err := constants.StringArrayToRoleArray(rolesAsString)

	userRole := constants.EnumRole.User

	if !slices.Contains(*roles, userRole) {
		*roles = append(*roles, userRole)
	}

	if len(requiredRoles) == 0 {
		requiredRoles = []constants.Role{userRole}
	}

	hasEnoughPermissions := false

	if len(requiredRoles) == 0 {
		hasEnoughPermissions = true // no roles required
	} else {
		for _, requiredRole := range requiredRoles {
			if slices.Contains(*roles, requiredRole) {
				hasEnoughPermissions = true
				break
			}
		}
	}

	if !hasEnoughPermissions {
		return nil, nil, constants.ErrorNotEnoughPermissions
	}

	return response.Claims, roles, nil

}

func AuthenticateAndGetId(ctx *fiber.Ctx, roles ...constants.Role) (string, error) {

	claims, _, err := authenticate(ctx, roles...)

	if err != nil {
		return "", err
	}

	id, ok := claims["id"].(string)

	if !ok {
		return "", constants.ErrorUserIdCast
	}

	return id, nil

}

func Authenticate(ctx *fiber.Ctx, roles ...constants.Role) (string, *[]constants.Role, error) {

	claims, userRoles, err := authenticate(ctx, roles...)

	if err != nil {
		return "", nil, err
	}

	id, ok := claims["id"].(string)

	if !ok {
		return "", nil, constants.ErrorUserIdCast
	}

	return id, userRoles, nil

}
