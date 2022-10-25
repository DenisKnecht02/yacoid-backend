package main

import (
	"fmt"
	"yacoid_server/api"
	"yacoid_server/common"
	"yacoid_server/database"
)

// TODO: Filter, Sort

func main() {

	err := common.LoadEnvironmentVariables()

	if err != nil {
		panic(fmt.Sprintf("Failed to load env variables: %v\n", err))
	}

	err = database.Connect()

	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v\n", err))
	}

	api.StartAPI()

}
