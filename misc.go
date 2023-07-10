package main

import (
	"fmt"
	"strings"

	"github.com/sunshineplan/database/mongodb"
)

func addUser(username string) error {
	svc.Print("Start!")
	if err := initDB(); err != nil {
		return err
	}

	username = strings.TrimSpace(strings.ToLower(username))

	insertedID, err := accountClient.InsertOne(
		struct {
			Username string `json:"username" bson:"username"`
			Password string `json:"password" bson:"password"`
			Uid      string `json:"uid" bson:"uid"`
		}{username, "123456", username},
	)
	if err != nil {
		return err
	}

	if _, err := addTask(task{
		Task: "Welcome to use mytasks!",
		List: "My Tasks",
	}, insertedID.(mongodb.ObjectID).Hex(), false); err != nil {
		return err
	}
	svc.Print("Done!")
	return nil
}

func deleteUser(username string) error {
	svc.Print("Start!")
	if err := initDB(); err != nil {
		return err
	}

	username = strings.TrimSpace(strings.ToLower(username))

	deletedCount, err := accountClient.DeleteOne(mongodb.M{"username": username})
	if err != nil {
		return err
	} else if deletedCount == 0 {
		return fmt.Errorf("user %s does not exist", username)
	}
	svc.Print("Done!")
	return nil
}
