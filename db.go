package main

import (
	"github.com/sunshineplan/database/mongodb/api"
	"github.com/sunshineplan/utils"
)

var accountClient api.Client
var incompleteClient api.Client
var completedClient api.Client

func initDB() error {
	var mongo api.Client
	if err := utils.Retry(func() error {
		return meta.Get("mytasks_mongo", &mongo)
	}, 3, 20); err != nil {
		return err
	}

	accountClient, incompleteClient, completedClient = mongo, mongo, mongo
	accountClient.Collection = "account"
	incompleteClient.Collection = "incomplete"
	completedClient.Collection = "completed"

	return nil
}

func test() (err error) {
	var mongo api.Client
	return meta.Get("mytasks_mongo", &mongo)
}
