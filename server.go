package main

import (
	"fmt"
	"yacoid_server/api"
	"yacoid_server/auth"
	"yacoid_server/common"
	"yacoid_server/database"
)

// TODO: Filter, Sort

// next steps: middleware with roles, setup environment in insomnia, check functionality of approving/rejecting, endpoints for pagination of authors etc.

func main() {

	err := common.LoadEnvironmentVariables()

	if err != nil {
		panic(fmt.Sprintf("Failed to load env variables: %v\n", err))
	}

	err = database.Connect()

	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v\n", err))
	}

	err = auth.Initialize()

	if err != nil {
		panic(fmt.Sprintf("Failed to connect to auth system: %v\n", err))
	}

	api.StartAPI()

}
