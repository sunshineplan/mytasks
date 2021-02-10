package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func checkIncomplete(objecdID primitive.ObjectID, userID interface{}) bool {
	return checkTask(objecdID, userID, false)
}

func addIncomplete(c *gin.Context) {
	var task task
	if err := c.BindJSON(&task); err != nil {
		c.String(400, "")
		return
	}

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	insertID, err := addTask(task, userID, false)
	if err != nil {
		log.Println("Failed to add incomplete task:", err)
		c.String(500, "")
		return
	}

	c.JSON(200, gin.H{"status": 1, "id": insertID})
	return
}

func editIncomplete(c *gin.Context) {
	objectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	var task task
	if err := c.BindJSON(&task); err != nil {
		c.String(400, "")
		return
	}
	task.Task = strings.TrimSpace(task.Task)

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	} else if checkIncomplete(objectID, userID) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err := collIncomplete.UpdateOne(
			ctx, bson.M{"_id": objectID}, bson.M{"$set": bson.M{"task": task.Task}}); err != nil {
			log.Println("Failed to edit incomplete task:", err)
			c.String(500, "")
			return
		}

		c.JSON(200, gin.H{"status": 1})
		return
	}

	c.String(403, "")
}

func completeTask(c *gin.Context) {
	objectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	} else if checkIncomplete(objectID, userID) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var task task
		if err := collIncomplete.FindOneAndDelete(ctx, bson.M{"_id": objectID}).Decode(&task); err != nil {
			log.Println("Failed to get incomplete task:", err)
			c.String(500, "")
			return
		}

		insertID, err := addTask(task, userID, true)
		if err != nil {
			log.Println("Failed to add completed task:", err)
			c.String(500, "")
			return
		}

		c.JSON(200, gin.H{"status": 1, "id": insertID})
		return
	}

	c.String(403, "")
}

func deleteIncomplete(c *gin.Context) {
	objectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	} else if checkIncomplete(objectID, userID) {
		if err := deleteTask(objectID, userID, false); err != nil {
			log.Println("Failed to delete completed task:", err)
			c.String(500, "")
			return
		}

		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.String(403, "")
}

func reorder(c *gin.Context) {
	var data struct{ List, Orig, Dest string }
	if err := c.BindJSON(&data); err != nil {
		c.String(400, "")
		return
	}

	orig, err := primitive.ObjectIDFromHex(data.Orig)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	dest, err := primitive.ObjectIDFromHex(data.Dest)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	} else if !checkIncomplete(orig, userID) || !checkIncomplete(dest, userID) {
		c.String(403, "")
		return
	}

	if err := reorderTask(userID, data.List, orig, dest); err != nil {
		log.Println("Failed to reorder tasks:", err)
		c.String(500, "")
		return
	}

	c.String(200, "1")
}
