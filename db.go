package main

import (
	"github.com/sunshineplan/database/mongodb"
	"github.com/sunshineplan/database/mongodb/api"
	"github.com/sunshineplan/utils"
)

var accountClient mongodb.Client
var incompleteClient mongodb.Client
var completedClient mongodb.Client

func initDB() (err error) {
	var apiClient api.Client
	if err = utils.Retry(func() error {
		return meta.Get("mytasks_mongo", &apiClient)
	}, 3, 20); err != nil {
		return err
	}

	account, incomplete, completed := apiClient, apiClient, apiClient
	account.Collection = "account"
	incomplete.Collection = "incomplete"
	completed.Collection = "completed"
	accountClient, incompleteClient, completedClient = &account, &incomplete, &completed

	if err = accountClient.Connect(); err != nil {
		return
	}
	if err = incompleteClient.Connect(); err != nil {
		return
	}
	return completedClient.Connect()
}

func test() (err error) {
	return initDB()
}
