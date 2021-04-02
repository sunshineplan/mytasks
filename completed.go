package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func checkCompleted(objecdID primitive.ObjectID, userID interface{}) bool {
	return checkTask(objecdID, userID, true)
}

func moreCompleted(c *gin.Context) {
	var data struct {
		List  string
		Start int64
	}
	if err := c.BindJSON(&data); err != nil {
		c.String(400, "")
		return
	}

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collCompleted.Find(ctx,
		bson.M{"list": data.List, "user": userID},
		options.Find().SetSort(bson.M{"created": 1}).SetLimit(30).SetSkip(data.Start),
	)
	if err != nil {
		log.Println("Failed to query tasks:", err)
		c.String(500, "")
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tasks := []task{}
	if err = cursor.All(ctx, &tasks); err != nil {
		log.Println("Failed to get tasks:", err)
		c.String(500, "")
		return
	}
	for i := range tasks {
		tasks[i].ID = tasks[i].ObjectID.Hex()
	}

	c.JSON(200, tasks)
}

func revertCompleted(c *gin.Context) {
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
	} else if checkCompleted(objectID, userID) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var task task
		if err := collCompleted.FindOneAndDelete(ctx, bson.M{"_id": objectID}).Decode(&task); err != nil {
			log.Println("Failed to get completed task:", err)
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

	c.String(403, "")
}

func deleteCompleted(c *gin.Context) {
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
	} else if checkCompleted(objectID, userID) {
		if err := deleteTask(objectID, userID, true); err != nil {
			log.Println("Failed to delete completed task:", err)
			c.String(500, "")
			return
		}

		c.JSON(200, gin.H{"status": 1})
		return
	}

	c.String(403, "")
}

func emptyCompleted(c *gin.Context) {
	var data struct{ List string }
	if err := c.BindJSON(&data); err != nil {
		log.Print(err)
		c.String(400, "")
		return
	}

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := collCompleted.DeleteMany(
		ctx, bson.M{"user": userID, "list": data.List}); err != nil {
		log.Println("Failed to empty completed tasks:", err)
		c.String(500, "")
		return
	}

	c.JSON(200, gin.H{"status": 1})
}
