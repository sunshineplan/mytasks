package main

import (
	"log"
	"strings"

	"github.com/sunshineplan/database/mongodb/api"
)

func addUser(username string) {
	log.Print("Start!")
	if err := initDB(); err != nil {
		log.Fatalln("Failed to initialize database:", err)
	}

	username = strings.TrimSpace(strings.ToLower(username))

	insertedID, err := accountClient.InsertOne(api.M{
		"username": username,
		"password": "123456",
		"uid":      username,
	})
	if err != nil {
		log.Fatal(err)
	}

	if _, err := addTask(task{
		Task: "Welcome to use mytasks!",
		List: "My Tasks",
	}, insertedID, false); err != nil {
		log.Fatal(err)
	}
	log.Print("Done!")
}

func deleteUser(username string) {
	log.Print("Start!")
	if err := initDB(); err != nil {
		log.Fatalln("Failed to initialize database:", err)
	}

	username = strings.TrimSpace(strings.ToLower(username))

	deletedCount, err := accountClient.DeleteOne(api.M{"username": username})
	if err != nil {
		log.Fatalln("Failed to delete user:", err)
	} else if deletedCount == 0 {
		log.Fatalf("User %s does not exist.", username)
	}
	log.Print("Done!")
}
