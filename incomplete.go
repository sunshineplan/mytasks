package main

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb"
)

func checkIncomplete(id mongodb.ObjectID, userID string) bool {
	return checkTask(id, userID, false)
}

func addIncomplete(c *gin.Context) {
	var task task
	if err := c.BindJSON(&task); err != nil {
		c.Status(400)
		return
	}
	if task.List == "" {
		c.Status(400)
		return
	}

	user, err := getUser(c)
	if err != nil {
		svc.Print(err)
		c.Status(500)
		return
	}

	id, seq, err := addTask(task, user.ID.Hex(), false)
	if err != nil {
		svc.Println("Failed to add incomplete task:", err)
		c.Status(500)
		return
	}
	newLastModified(user.ID.Hex(), c)
	c.JSON(200, gin.H{"status": 1, "id": id.Hex(), "seq": seq})
}

func editIncomplete(c *gin.Context) {
	id, err := incompleteClient.ObjectID(c.Param("id"))
	if err != nil {
		svc.Print(err)
		c.Status(400)
		return
	}

	var task struct{ Task string }
	if err := c.BindJSON(&task); err != nil {
		c.Status(400)
		return
	}
	task.Task = strings.TrimSpace(task.Task)

	user, err := getUser(c)
	if err != nil {
		svc.Print(err)
		c.Status(500)
		return
	} else if checkIncomplete(id, user.ID.Hex()) {
		if _, err := incompleteClient.UpdateOne(
			mongodb.M{"_id": id},
			mongodb.M{"$set": mongodb.M{"task": task.Task}},
			nil,
		); err != nil {
			svc.Println("Failed to edit incomplete task:", err)
			c.Status(500)
			return
		}
		newLastModified(user.ID.Hex(), c)
		c.JSON(200, gin.H{"status": 1})
		return
	}

	c.String(403, "")
}

func completeTask(c *gin.Context) {
	id, err := incompleteClient.ObjectID(c.Param("id"))
	if err != nil {
		svc.Print(err)
		c.Status(400)
		return
	}
	user, err := getUser(c)
	if err != nil {
		svc.Print(err)
		c.Status(500)
		return
	} else if checkIncomplete(id, user.ID.Hex()) {
		var task task
		if err := incompleteClient.FindOneAndDelete(mongodb.M{"_id": id}, nil, &task); err != nil {
			svc.Println("Failed to get incomplete task:", err)
			c.Status(500)
			return
		}

		id, _, err := addTask(task, user.ID.Hex(), true)
		if err != nil {
			svc.Println("Failed to add completed task:", err)
			c.Status(500)
			return
		}
		newLastModified(user.ID.Hex(), c)
		c.JSON(200, gin.H{"status": 1, "id": id.Hex()})
		return
	}

	c.String(403, "")
}

func deleteIncomplete(c *gin.Context) {
	id, err := incompleteClient.ObjectID(c.Param("id"))
	if err != nil {
		svc.Print(err)
		c.Status(400)
		return
	}
	user, err := getUser(c)
	if err != nil {
		svc.Print(err)
		c.Status(500)
		return
	} else if checkIncomplete(id, user.ID.Hex()) {
		if err := deleteTask(id, user.ID.Hex(), false); err != nil {
			svc.Println("Failed to delete completed task:", err)
			c.JSON(200, gin.H{"status": 0})
			return
		}
		newLastModified(user.ID.Hex(), c)
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.String(403, "")
}

func reorder(c *gin.Context) {
	var data struct{ List, Orig, Dest string }
	if err := c.BindJSON(&data); err != nil {
		c.Status(400)
		return
	}

	origID, err := incompleteClient.ObjectID(data.Orig)
	if err != nil {
		svc.Print(err)
		c.Status(400)
		return
	}
	destID, err := incompleteClient.ObjectID(data.Dest)
	if err != nil {
		svc.Print(err)
		c.Status(400)
		return
	}

	user, err := getUser(c)
	if err != nil {
		svc.Print(err)
		c.Status(500)
		return
	} else if !checkIncomplete(origID, user.ID.Hex()) || !checkIncomplete(destID, user.ID.Hex()) {
		c.String(403, "")
		return
	}

	if err := reorderTask(user.ID.Hex(), data.List, origID, destID); err != nil {
		svc.Println("Failed to reorder tasks:", err)
		c.Status(500)
		return
	}
	newLastModified(user.ID.Hex(), c)
	c.String(200, "1")
}
