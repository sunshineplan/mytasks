package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb"
)

func checkCompleted(id mongodb.ObjectID, userID string) bool {
	return checkTask(id, userID, true)
}

func moreCompleted(c *gin.Context) {
	var data struct {
		List  string
		Start int64
	}
	if err := c.BindJSON(&data); err != nil {
		c.Status(400)
		return
	}

	user, err := getUser(c)
	if err != nil {
		svc.Print(err)
		c.Status(500)
		return
	}

	tasks := []task{}
	if err := completedClient.Find(
		mongodb.M{"list": data.List, "user": user.ID.Hex()},
		&mongodb.FindOpt{Sort: mongodb.M{"created": -1}, Limit: 30, Skip: data.Start},
		&tasks,
	); err != nil {
		svc.Println("Failed to query tasks:", err)
		c.Status(500)
		return
	}
	for i := range tasks {
		tasks[i].ID = tasks[i].ObjectID.Hex()
		tasks[i].ObjectID = nil
	}

	c.JSON(200, tasks)
}

func revertCompleted(c *gin.Context) {
	id, err := completedClient.ObjectID(c.Param("id"))
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
	} else if checkCompleted(id, user.ID.Hex()) {
		var task task
		if err := completedClient.FindOneAndDelete(mongodb.M{"_id": id}, nil, &task); err != nil {
			svc.Println("Failed to get completed task:", err)
			c.Status(500)
			return
		}

		insertID, seq, err := addTask(task, user.ID.Hex(), false)
		if err != nil {
			svc.Println("Failed to add incomplete task:", err)
			c.Status(500)
			return
		}
		newLastModified(user.ID.Hex(), c)
		c.JSON(200, gin.H{"status": 1, "id": insertID.Hex(), "seq": seq})
		return
	}

	c.String(403, "")
}

func deleteCompleted(c *gin.Context) {
	id, err := completedClient.ObjectID(c.Param("id"))
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
	} else if checkCompleted(id, user.ID.Hex()) {
		if err := deleteTask(id, user.ID.Hex(), true); err != nil {
			svc.Println("Failed to delete completed task:", err)
			c.Status(500)
			return
		}
		newLastModified(user.ID.Hex(), c)
		c.JSON(200, gin.H{"status": 1})
		return
	}

	c.String(403, "")
}

func emptyCompleted(c *gin.Context) {
	var data struct{ List string }
	if err := c.BindJSON(&data); err != nil {
		svc.Print(err)
		c.Status(400)
		return
	}

	user, err := getUser(c)
	if err != nil {
		svc.Print(err)
		c.Status(500)
		return
	}

	if _, err := completedClient.DeleteMany(mongodb.M{"user": user.ID, "list": data.List}); err != nil {
		svc.Println("Failed to empty completed tasks:", err)
		c.Status(500)
		return
	}
	newLastModified(user.ID.Hex(), c)
	c.JSON(200, gin.H{"status": 1})
}
