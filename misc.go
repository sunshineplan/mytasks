package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func addUser(username string) {
	log.Print("Start!")
	if err := initDB(); err != nil {
		log.Fatalln("Failed to initialize database:", err)
	}

	username = strings.TrimSpace(strings.ToLower(username))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := collAccount.InsertOne(ctx, bson.D{
		{Key: "username", Value: username},
		{Key: "password", Value: "123456"},
		{Key: "uid", Value: username},
	})
	if err != nil {
		log.Fatal(err)
	}
	objectID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		log.Fatal("Failed to get last insert id.")
	}
	if _, err := collIncomplete.InsertOne(ctx, bson.D{
		{Key: "task", Value: "Welcome to use mytasks!"},
		{Key: "list", Value: "My Tasks"},
		{Key: "created", Value: time.Now()},
		{Key: "user", Value: objectID.Hex()},
		{Key: "seq", Value: 1},
	}); err != nil {
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := collAccount.DeleteOne(ctx, bson.M{"username": username})
	if err != nil {
		log.Fatalln("Failed to delete user:", err)
	} else if res.DeletedCount == 0 {
		log.Fatalf("User %s does not exist.", username)
	}
	log.Print("Done!")
}

func backup(file string) {
	log.Print("Start!")
	if err := initDB(); err != nil {
		log.Fatalln("Failed to initialize database:", err)
	}

	if err := dbConfig.Backup(file); err != nil {
		log.Fatal(err)
	}
	log.Print("Done!")
}

func restore(file string) {
	log.Print("Start!")
	if _, err := os.Stat(file); err != nil {
		log.Fatalln("File not found:", err)
	}

	if err := initDB(); err != nil {
		log.Fatalln("Failed to initialize database:", err)
	}

	if err := dbConfig.Restore(file); err != nil {
		log.Fatal(err)
	}
	log.Print("Done!")
}
