package main

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb"
)

type list struct {
	List       string `json:"list"`
	Incomplete int    `json:"incomplete"`
	Completed  int    `json:"completed"`
}

func getList(c *gin.Context) {
	userID, _ := c.Get("id")
	lists := []list{}
	var incomplete, completed []struct {
		List  string `json:"_id"`
		Count int
	}
	ec := make(chan error, 1)
	go func() {
		ec <- incompleteClient.Aggregate([]mongodb.M{
			{"$match": mongodb.M{"user": userID}},
			{"$group": mongodb.M{"_id": "$list", "count": mongodb.M{"$sum": 1}}},
			{"$sort": mongodb.M{"count": 1}},
		}, &incomplete)
	}()

	if err := completedClient.Aggregate([]mongodb.M{
		{"$match": mongodb.M{"user": userID}},
		{"$group": mongodb.M{"_id": "$list", "count": mongodb.M{"$sum": 1}}},
	}, &completed); err != nil {
		svc.Println("Failed to get completed tasks:", err)
		c.Status(500)
		return
	}

	if err := <-ec; err != nil {
		svc.Println("Failed to incomplete tasks:", err)
		c.Status(500)
		return
	}

	for _, i := range incomplete {
		lists = append(lists, list{List: i.List, Incomplete: i.Count})
	}
Loop:
	for _, i := range completed {
		for index := range lists {
			if lists[index].List == i.List {
				lists[index].Completed = i.Count
				continue Loop
			}
		}
		lists = append(lists, list{List: i.List, Completed: i.Count})
	}

	c.JSON(200, lists)
}

func editList(c *gin.Context) {
	var data struct{ Old, New string }
	if err := c.BindJSON(&data); err != nil {
		svc.Print(err)
		c.Status(400)
		return
	}
	data.New = strings.TrimSpace(data.New)

	userID, _ := c.Get("id")

	var message string
	exist, err := checkExist(mongodb.M{"user": userID, "list": data.New})
	switch {
	case data.New == "":
		message = "New list name is empty."
	case data.Old == data.New:
		message = "New list name is same as old list."
	case len(data.New) > 15:
		message = "List name exceeded length limit."
	case err != nil:
		svc.Println("Failed to get list:", err)
		c.Status(500)
		return
	case exist:
		message = fmt.Sprintf("List %s is already existed.", data.New)
	default:
		ec := make(chan error, 1)
		go func() {
			_, err := incompleteClient.UpdateMany(
				mongodb.M{"user": userID, "list": data.Old},
				mongodb.M{"$set": mongodb.M{"list": data.New}},
				nil,
			)
			ec <- err
		}()
		if _, err := completedClient.UpdateMany(
			mongodb.M{"user": userID, "list": data.Old},
			mongodb.M{"$set": mongodb.M{"list": data.New}},
			nil,
		); err != nil {
			svc.Println("Failed to edit completed tasks list:", err)
			c.Status(500)
			return
		}

		if err := <-ec; err != nil {
			svc.Println("Failed to edit incomplete tasks list:", err)
			c.Status(500)
			return
		}
		newLastModified(userID, c)
		c.JSON(200, gin.H{"status": 1})
		return
	}

	c.JSON(200, gin.H{"status": 0, "message": message})
}

func deleteList(c *gin.Context) {
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

	ec := make(chan error, 1)
	go func() {
		_, err := incompleteClient.DeleteMany(mongodb.M{"user": user.ID, "list": data.List})
		ec <- err
	}()
	if _, err := completedClient.DeleteMany(mongodb.M{"user": user.ID, "list": data.List}); err != nil {
		svc.Println("Failed to delete completed tasks list:", err)
		c.JSON(200, gin.H{"status": 0})
		return
	}

	if err := <-ec; err != nil {
		svc.Println("Failed to delete incomplete tasks list:", err)
		c.JSON(200, gin.H{"status": 0})
		return
	}
	newLastModified(user.ID, c)
	c.JSON(200, gin.H{"status": 1})
}
