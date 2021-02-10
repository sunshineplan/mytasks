package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type list struct {
	List       string `json:"list"`
	Incomplete int    `json:"incomplete"`
	Completed  int    `json:"completed"`
}

func getList(userID string) ([]list, error) {
	lists := []list{}
	var incomplete, completed []struct {
		List  string `bson:"_id"`
		Count int
	}
	c := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := collIncomplete.Aggregate(ctx, []bson.M{
			{"$match": bson.M{"user": userID}},
			{"$group": bson.M{"_id": "$list", "count": bson.M{"$sum": 1}}},
			{"$sort": bson.M{"count": 1}},
		})
		if err != nil {
			log.Println("Failed to query incomplete tasks:", err)
			return
		}

		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		c <- cursor.All(ctx, &incomplete)
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collCompleted.Aggregate(ctx, []bson.M{
		{"$match": bson.M{"user": userID}},
		{"$group": bson.M{"_id": "$list", "count": bson.M{"$sum": 1}}},
	})
	if err != nil {
		log.Println("Failed to query completed tasks:", err)
		return lists, err
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := cursor.All(ctx, &completed); err != nil {
		log.Println("Failed to get categories:", err)
		return lists, err
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

	return lists, nil
}

func editList(c *gin.Context) {
	var data struct{ Old, New string }
	if err := c.BindJSON(&data); err != nil {
		log.Print(err)
		c.String(400, "")
		return
	}
	data.New = strings.TrimSpace(data.New)

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	lists, err := getList(userID)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	var exist bool
	for _, i := range lists {
		if i.List == data.New {
			exist = true
		}
	}

	var message string
	switch {
	case data.New == "":
		message = "New list name is empty."
	case data.Old == data.New:
		message = "New list name is same as old list."
	case len(data.New) > 15:
		message = "List name exceeded length limit."
	case exist:
		message = fmt.Sprintf("List %s is already existed.", data.New)
	default:
		ec := make(chan error, 1)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_, err := collIncomplete.UpdateMany(ctx,
				bson.M{"user": userID, "list": data.Old},
				bson.M{"$set": bson.M{"list": data.New}},
			)

			ec <- err
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err := collCompleted.UpdateMany(ctx,
			bson.M{"user": userID, "list": data.Old},
			bson.M{"$set": bson.M{"list": data.New}},
		); err != nil {
			log.Println("Failed to edit completed tasks list:", err)
			c.String(500, "")
			return
		}

		if err := <-ec; err != nil {
			log.Println("Failed to edit incomplete tasks list:", err)
			c.String(500, "")
			return
		}

		c.JSON(200, gin.H{"status": 1})
		return
	}

	c.JSON(200, gin.H{"status": 0, "message": message})
}

func deleteList(c *gin.Context) {
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

	ec := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := collIncomplete.DeleteMany(
			ctx, bson.M{"user": userID, "list": data.List})

		ec <- err
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := collCompleted.DeleteMany(
		ctx, bson.M{"user": userID, "list": data.List}); err != nil {
		log.Println("Failed to delete completed tasks list:", err)
		c.String(500, "")
		return
	}

	if err := <-ec; err != nil {
		log.Println("Failed to delete incomplete tasks list:", err)
		c.String(500, "")
		return
	}

	c.JSON(200, gin.H{"status": 1})
	return
}
