package auth

import (
	"os"
	"yacoid_server/constants"

	"github.com/authorizerdev/authorizer-go"
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
