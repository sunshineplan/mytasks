package main

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb/api"
)

func checkIncomplete(id string, userID interface{}) bool {
	return checkTask(id, userID, false)
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
}

func editIncomplete(c *gin.Context) {
	id := c.Param("id")

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
	} else if checkIncomplete(id, userID) {
		if _, err := incompleteClient.UpdateOne(
			api.M{"_id": api.ObjectID(id)},
			api.M{"$set": api.M{"task": task.Task}},
			nil,
		); err != nil {
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
	id := c.Param("id")
	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	} else if checkIncomplete(id, userID) {
		var task task
		if err := incompleteClient.FindOneAndDelete(api.M{"_id": api.ObjectID(id)}, nil, &task); err != nil {
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
	id := c.Param("id")
	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	} else if checkIncomplete(id, userID) {
		if err := deleteTask(id, userID, false); err != nil {
			log.Println("Failed to delete completed task:", err)
			c.JSON(200, gin.H{"status": 0})
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

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	} else if !checkIncomplete(data.Orig, userID) || !checkIncomplete(data.Dest, userID) {
		c.String(403, "")
		return
	}

	if err := reorderTask(userID, data.List, data.Orig, data.Dest); err != nil {
		log.Println("Failed to reorder tasks:", err)
		c.String(500, "")
		return
	}

	c.String(200, "1")
}
