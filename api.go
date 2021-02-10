package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func info(c *gin.Context) {
	info := gin.H{}

	id, username, _ := getUser(c)
	if username == "" {
		c.JSON(200, info)
		return
	}
	info["username"] = username

	lists, err := getList(id)
	if err != nil {
		log.Println("Failed to get list:", err)
		c.String(500, "")
		return
	}
	info["lists"] = lists

	c.JSON(200, info)
}

func get(c *gin.Context) {
	var data task
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

	incomplete := []task{}
	completed := []task{}

	ec := make(chan error, 1)
	go func() {
		var err error
		incomplete, err = getTask(data.List, userID, false)
		ec <- err
	}()

	completed, err = getTask(data.List, userID, true)
	if err != nil {
		log.Println("Failed to get completed tasks:", err)
		c.String(500, "")
		return
	}

	if err := <-ec; err != nil {
		log.Println("Failed to get incomplete tasks:", err)
		c.String(500, "")
		return
	}

	c.JSON(200, gin.H{"incomplete": incomplete, "completed": completed})
}
