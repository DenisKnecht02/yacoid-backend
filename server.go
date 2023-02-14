package main

import (
	"fmt"
	"yacoid_server/api"
	"yacoid_server/auth"
	"yacoid_server/common"
	"yacoid_server/database"
)

func main() {

	err := common.LoadEnvironmentVariables()

	if err != nil {
		fmt.Printf("Failed to load env variables from .env-file. Using OS env variables. Error: %v\n", err)
	}

	err = database.Connect()

	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v\n", err))
	}

	err = auth.Initialize()

	if err != nil {
		panic(fmt.Sprintf("Failed to connect to auth system: %v\n", err))
	}

	err = api.StartAPI()
	
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to auth system: %v\n", err))
	}

}
