package main

import (
	"log"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func info(c *gin.Context) {
	info := gin.H{}
	userID := sessions.Default(c).Get("userID")
	if userID == nil {
		c.JSON(200, info)
		return
	}

	var username string
	if err := db.QueryRow("SELECT username FROM user WHERE id = ?", userID).Scan(&username); err != nil {
		log.Println("Failed to get username:", err)
		c.String(500, "")
		return
	}
	info["username"] = username

	lists, err := getList(userID)
	if err != nil {
		log.Println("Failed to get list:", err)
		c.String(500, "")
		return
	}
	info["lists"] = lists

	c.JSON(200, info)
}

func get(c *gin.Context) {
	var option task
	if err := c.BindJSON(&option); err != nil {
		c.String(400, "")
		return
	}

	userID := sessions.Default(c).Get("userID")
	if !checkList(option.List, userID) {
		c.String(403, "")
		return
	}

	incomplete := []task{}
	completed := []task{}

	ec := make(chan error, 1)
	go func() {
		rows, err := db.Query(
			"SELECT task_id, task, list_id, created FROM tasks WHERE list_id = ? AND user_id = ?",
			option.List, userID)
		if err != nil {
			ec <- err
			return
		}
		defer rows.Close()
		for rows.Next() {
			var task task
			if err := rows.Scan(&task.ID, &task.Task, &task.List, &task.Created); err != nil {
				ec <- err
				return
			}
			incomplete = append(incomplete, task)
		}
		ec <- nil
	}()
	rows, err := db.Query(
		"SELECT task_id, task, list_id, created FROM completeds WHERE list_id = ? AND user_id = ? LIMIT 10",
		option.List, userID)
	if err != nil {
		log.Println("Failed to get completeds:", err)
		c.String(500, "")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var task task
		if err := rows.Scan(&task.ID, &task.Task, &task.List, &task.Created); err != nil {
			log.Println("Failed to scan completeds:", err)
			c.String(500, "")
			return
		}
		completed = append(completed, task)
	}
	if err := <-ec; err != nil {
		log.Println("Failed to get tasks:", err)
		c.String(500, "")
		return
	}

	c.JSON(200, gin.H{"incomplete": incomplete, "completed": completed})
}
