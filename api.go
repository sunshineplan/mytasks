package main

import (
	"github.com/gin-gonic/gin"
)

func info(c *gin.Context) {
	user, _ := getUser(c)
	if user.Username == "" {
		c.JSON(200, struct{}{})
		return
	}
	c.Set("last", user.Last)
	last, ok := checkLastModified(c)
	c.SetCookie("last", last, 856400*365, "", "", false, false)
	if ok {
		lists, err := getList(user.ID)
		if err != nil {
			svc.Println("Failed to get list:", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"username": user.Username, "lists": lists})
	} else {
		c.Status(409)
	}
}

func get(c *gin.Context) {
	var data task
	if err := c.BindJSON(&data); err != nil {
		c.String(400, "")
		return
	}

	user, err := getUser(c)
	if err != nil {
		svc.Print(err)
		c.String(500, "")
		return
	}

	var incomplete []task
	ec := make(chan error, 1)
	go func() {
		var err error
		incomplete, err = getTask(data.List, user.ID, false)
		ec <- err
	}()

	completed, err := getTask(data.List, user.ID, true)
	if err != nil {
		svc.Println("Failed to get completed tasks:", err)
		c.String(500, "")
		return
	}

	if err := <-ec; err != nil {
		svc.Println("Failed to get incomplete tasks:", err)
		c.String(500, "")
		return
	}

	c.JSON(200, gin.H{"incomplete": incomplete, "completed": completed})
}
