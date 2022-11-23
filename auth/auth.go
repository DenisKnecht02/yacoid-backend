package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"yacoid_server/common"
	"yacoid_server/constants"

	"github.com/authorizerdev/authorizer-go"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var AuthClient *authorizer.AuthorizerClient

func Initialize() error {

	defaultHeaders := map[string]string{}

	var err error
	AuthClient, err = authorizer.NewAuthorizerClient(os.Getenv(constants.AUTH_CLIENT_ID), os.Getenv(constants.AUTH_URL), os.Getenv(constants.AUTH_REDIRECT_URL), defaultHeaders)

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

// Can be the id or access token
func GetAuthorizationToken(ctx *fiber.Ctx) (string, error) {

	authorizationHeader := ctx.GetReqHeaders()["Authorization"]
	tokenSplit := strings.Split(authorizationHeader, "Bearer ")

	if len(tokenSplit) < 2 || tokenSplit[1] == "" {
		return "", fiber.ErrUnauthorized
	}

	return tokenSplit[1], nil

}

func GetIdTokenAndExpectRoleFromContext(ctx *fiber.Ctx, role constants.Role) (string, error) {

	idToken, err := GetAuthorizationToken(ctx)

	if err != nil {
		return "", err
	}

	response, err := AuthClient.ValidateJWTToken(&authorizer.ValidateJWTTokenInput{
		TokenType: authorizer.TokenTypeIDToken,
		Token:     idToken,
		Roles:     []*string{role.StringAddress()},
	})

	if err != nil {
		return "", err
	}

	if !response.IsValid {
		return "", common.ErrorValidationResponseInvalid
	}

	return idToken, common.ErrorMissingRole

}

func GetUserByContext(ctx *fiber.Ctx) (*authorizer.User, error) {

	token, err := GetAuthorizationToken(ctx)

	if err != nil {
		return nil, err
	}

	return GetUserByToken(token)
}

func DecodeJWTToken(bearer string) (*jwt.Token, *jwt.MapClaims, error) {

	tokenString := strings.Split(bearer, "Bearer ")[1]

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, common.ErrorUnexpectedSigningMethod
		}

		key, err := GetPublicKey()
		fmt.Println(key)

		if err != nil {
			return nil, err
		}

		return key, nil
	})

	if err != nil {
		return nil, nil, err
	}

	return token, &claims, nil

}

func DecodeJWTTokenWithContext(ctx *fiber.Ctx) (*jwt.Token, *jwt.MapClaims, error) {
	return DecodeJWTToken(ctx.GetReqHeaders()["Authorization"])
}

type JWTConfig struct {
	Type string `json:"type"`
	Key  string `json:"key"`
}

func GetPublicKey() ([]byte, error) {

	content, err := ioutil.ReadFile("jwtConfig.json")
	key := []byte{}

	if err != nil {
		return key, err
	}

	var config JWTConfig
	err = json.Unmarshal(content, &config)

	if err != nil {
		return key, err
	}

	key = []byte(config.Key)
	return key, nil

}
