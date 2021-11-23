package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb/api"
)

func checkCompleted(id string, userID interface{}) bool {
	return checkTask(id, userID, true)
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

	tasks := []task{}
	if err := completedClient.Find(
		api.M{"list": data.List, "user": userID},
		&api.FindOpt{Sort: api.M{"created": 1}, Limit: 30, Skip: data.Start},
		&tasks,
	); err != nil {
		log.Println("Failed to query tasks:", err)
		c.String(500, "")
		return
	}
	for i := range tasks {
		tasks[i].ID = tasks[i].ObjectID
		tasks[i].ObjectID = ""
	}

	c.JSON(200, tasks)
}

func revertCompleted(c *gin.Context) {
	id := c.Param("id")
	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	} else if checkCompleted(id, userID) {
		var task task
		if err := completedClient.FindOneAndDelete(api.M{"_id": api.ObjectID(id)}, nil, &task); err != nil {
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
	id := c.Param("id")
	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	} else if checkCompleted(id, userID) {
		if err := deleteTask(id, userID, true); err != nil {
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

	if _, err := completedClient.DeleteMany(api.M{"user": userID, "list": data.List}); err != nil {
		log.Println("Failed to empty completed tasks:", err)
		c.String(500, "")
		return
	}

	c.JSON(200, gin.H{"status": 1})
}
