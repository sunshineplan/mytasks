package main

import (
	"log"
	"strconv"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func moreCompleted(c *gin.Context) {
	var option struct{ List, Start int }
	if err := c.BindJSON(&option); err != nil {
		c.String(400, "")
		return
	}

	completed := []task{}
	rows, err := db.Query(
		"SELECT task_id, task, list_id, created FROM completeds WHERE list_id = ? AND user_id = ? LIMIT ?, 30",
		option.List, sessions.Default(c).Get("userID"), option.Start)
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
	c.JSON(200, completed)
}

func revertCompleted(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("Failed to get id param:", err)
		c.String(400, "")
		return
	}

	var exist string
	if err := db.QueryRow("SELECT task FROM completeds WHERE task_id = ? AND user_id = ?",
		id, sessions.Default(c).Get("userID")).Scan(&exist); err == nil {
		var insertID int
		if err := db.QueryRow("CALL revert_completed(?)", id).Scan(&insertID); err != nil {
			log.Println("Failed to revert completed task:", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1, "id": insertID})
		return
	}
	c.String(403, "")
}

func deleteCompleted(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("Failed to get id param:", err)
		c.String(400, "")
		return
	}

	var exist string
	if err := db.QueryRow("SELECT task FROM completeds WHERE task_id = ? AND user_id = ?",
		id, sessions.Default(c).Get("userID")).Scan(&exist); err == nil {
		if _, err := db.Exec("DELETE FROM completed WHERE id = ?", id); err != nil {
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("Failed to get id param:", err)
		c.String(400, "")
		return
	}

	var exist string
	if err := db.QueryRow("SELECT task FROM completeds WHERE list_id = ? AND user_id = ?",
		id, sessions.Default(c).Get("userID")).Scan(&exist); err == nil {
		if _, err := db.Exec("DELETE FROM completed WHERE list_id = ?", id); err != nil {
			log.Println("Failed to empty completed task:", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.String(403, "")
}
