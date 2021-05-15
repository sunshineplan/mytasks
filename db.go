package main

import (
	"github.com/sunshineplan/utils"
	"github.com/sunshineplan/utils/database/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

var dbConfig mongodb.Config
var collAccount *mongo.Collection
var collIncomplete *mongo.Collection
var collCompleted *mongo.Collection

func initDB() (err error) {
	if err = utils.Retry(func() error {
		return meta.Get("mytasks_mongo", &dbConfig)
	}, 3, 20); err != nil {
		return
	}

	var client *mongo.Client
	client, err = dbConfig.Open()
	if err != nil {
		return
	}

	database := client.Database(dbConfig.Database)

	collAccount = database.Collection("account")
	collIncomplete = database.Collection("incomplete")
	collCompleted = database.Collection("completed")

	return
}
