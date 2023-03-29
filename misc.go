package main

import (
	"strings"

	"github.com/sunshineplan/database/mongodb"
)

func addUser(username string) {
	svc.Print("Start!")
	if err := initDB(); err != nil {
		svc.Fatalln("Failed to initialize database:", err)
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
		svc.Fatal(err)
	}

	if _, err := addTask(task{
		Task: "Welcome to use mytasks!",
		List: "My Tasks",
	}, insertedID.(mongodb.ObjectID).Hex(), false); err != nil {
		svc.Fatal(err)
	}
	svc.Print("Done!")
}

func deleteUser(username string) {
	svc.Print("Start!")
	if err := initDB(); err != nil {
		svc.Fatalln("Failed to initialize database:", err)
	}

	username = strings.TrimSpace(strings.ToLower(username))

	deletedCount, err := accountClient.DeleteOne(mongodb.M{"username": username})
	if err != nil {
		svc.Fatalln("Failed to delete user:", err)
	} else if deletedCount == 0 {
		svc.Fatalf("User %s does not exist.", username)
	}
	svc.Print("Done!")
}
