package main

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb"
)

func checkIncomplete(id mongodb.ObjectID, userID any) bool {
	return checkTask(id, userID, false)
}

func addIncomplete(c *gin.Context) {
	var task task
	if err := c.BindJSON(&task); err != nil {
		c.String(400, "")
		return
	}

	user, err := getUser(c)
	if err != nil {
		svc.Print(err)
		c.String(500, "")
		return
	}

	insertID, err := addTask(task, user.ID, false)
	if err != nil {
		svc.Println("Failed to add incomplete task:", err)
		c.String(500, "")
		return
	}
	newLastModified(user.ID, c)
	c.JSON(200, gin.H{"status": 1, "id": insertID.(mongodb.ObjectID).Hex()})
}

func editIncomplete(c *gin.Context) {
	id, err := incompleteClient.ObjectID(c.Param("id"))
	if err != nil {
		svc.Print(err)
		c.String(400, "")
		return
	}

	var task task
	if err := c.BindJSON(&task); err != nil {
		c.String(400, "")
		return
	}
	task.Task = strings.TrimSpace(task.Task)

	user, err := getUser(c)
	if err != nil {
		svc.Print(err)
		c.String(500, "")
		return
	} else if checkIncomplete(id, user.ID) {
		if _, err := incompleteClient.UpdateOne(
			mongodb.M{"_id": id.Interface()},
			mongodb.M{"$set": mongodb.M{"task": task.Task}},
			nil,
		); err != nil {
			svc.Println("Failed to edit incomplete task:", err)
			c.String(500, "")
			return
		}
		newLastModified(user.ID, c)
		c.JSON(200, gin.H{"status": 1})
		return
	}

	c.String(403, "")
}

func completeTask(c *gin.Context) {
	id, err := incompleteClient.ObjectID(c.Param("id"))
	if err != nil {
		svc.Print(err)
		c.String(400, "")
		return
	}
	user, err := getUser(c)
	if err != nil {
		svc.Print(err)
		c.String(500, "")
		return
	} else if checkIncomplete(id, user.ID) {
		var task task
		if err := incompleteClient.FindOneAndDelete(mongodb.M{"_id": id.Interface()}, nil, &task); err != nil {
			svc.Println("Failed to get incomplete task:", err)
			c.String(500, "")
			return
		}

		insertID, err := addTask(task, user.ID, true)
		if err != nil {
			svc.Println("Failed to add completed task:", err)
			c.String(500, "")
			return
		}
		newLastModified(user.ID, c)
		c.JSON(200, gin.H{"status": 1, "id": insertID.(mongodb.ObjectID).Hex()})
		return
	}

	c.String(403, "")
}

func deleteIncomplete(c *gin.Context) {
	id, err := incompleteClient.ObjectID(c.Param("id"))
	if err != nil {
		svc.Print(err)
		c.String(400, "")
		return
	}
	user, err := getUser(c)
	if err != nil {
		svc.Print(err)
		c.String(500, "")
		return
	} else if checkIncomplete(id, user.ID) {
		if err := deleteTask(id, user.ID, false); err != nil {
			svc.Println("Failed to delete completed task:", err)
			c.JSON(200, gin.H{"status": 0})
			return
		}
		newLastModified(user.ID, c)
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

	origID, err := incompleteClient.ObjectID(data.Orig)
	if err != nil {
		svc.Print(err)
		c.String(400, "")
		return
	}
	destID, err := incompleteClient.ObjectID(data.Dest)
	if err != nil {
		svc.Print(err)
		c.String(400, "")
		return
	}

	user, err := getUser(c)
	if err != nil {
		svc.Print(err)
		c.String(500, "")
		return
	} else if !checkIncomplete(origID, user.ID) || !checkIncomplete(destID, user.ID) {
		c.String(403, "")
		return
	}

	if err := reorderTask(user.ID, data.List, origID, destID); err != nil {
		svc.Println("Failed to reorder tasks:", err)
		c.String(500, "")
		return
	}
	newLastModified(user.ID, c)
	c.String(200, "1")
}
