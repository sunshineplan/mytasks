package main

import (
	"time"

	"github.com/sunshineplan/database/mongodb"
	"github.com/sunshineplan/database/mongodb/driver"
	"github.com/sunshineplan/utils/retry"
)

var accountClient mongodb.Client
var incompleteClient mongodb.Client
var completedClient mongodb.Client

func initDB() (err error) {
	var client driver.Client
	if err = retry.Do(func() error {
		return meta.Get("mytasks_mongo", &client)
	}, 3, 20*time.Second); err != nil {
		return err
	}

	account, incomplete, completed := client, client, client
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

func test() error {
	return initDB()
}
