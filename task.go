package main

import (
	"log"
	"strconv"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type task struct {
	ID   int    `json:"id"`
	Task string `json:"task"`
	List int    `json:"list"`
}

func checkTask(taskID, userID interface{}) bool {
	var exist string
	if err := db.QueryRow("SELECT task FROM tasks WHERE task_id = ? AND user_id = ?",
		taskID, userID).Scan(&exist); err == nil {
		return true
	}
	return false
}

func checkAll(taskID, listID, userID interface{}) bool {
	var exist string
	if err := db.QueryRow("SELECT task FROM tasks WHERE task_id = ? AND list_id = ? AND user_id = ?",
		taskID, listID, userID).Scan(&exist); err == nil {
		return true
	}
	return false
}

func getTask(c *gin.Context) {
	var r task
	if err := c.BindJSON(&r); err != nil {
		c.String(400, "")
		return
	}

	rows, err := db.Query("SELECT task_id, task, list FROM tasks WHERE list_id = ? AND user_id = ?",
		r.List, sessions.Default(c).Get("userID"))
	if err != nil {
		log.Println("Failed to get tasks:", err)
		c.String(500, "")
		return
	}
	defer rows.Close()
	tasks := []task{}
	for rows.Next() {
		var task task
		if err := rows.Scan(&task.ID, &task.Task, &task.List); err != nil {
			log.Println("Failed to scan tasks:", err)
			c.String(500, "")
			return
		}
		tasks = append(tasks, task)
	}
	c.JSON(200, tasks)
}

func addTask(c *gin.Context) {
	var task task
	if err := c.BindJSON(&task); err != nil {
		c.String(400, "")
		return
	}

	if checkList(task.List, sessions.Default(c).Get("userID")) {
		if task.Task == "" {
			c.JSON(200, gin.H{"status": 0, "message": "Task is empty."})
			return
		}
		if _, err := db.Exec("INSERT INTO task (task, list_id) VALUES (?, ?)",
			task.Task, task.List); err != nil {
			log.Println("Failed to add task:", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.String(403, "")
}

func editTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("Failed to get id param:", err)
		c.String(400, "")
		return
	}
	var task task
	if err := c.BindJSON(&task); err != nil {
		c.String(400, "")
		return
	}

	if checkAll(id, task.List, sessions.Default(c).Get("userID")) {
		if task.Task == "" {
			c.JSON(200, gin.H{"status": 0, "message": "Task is empty."})
			return
		}
		if _, err := db.Exec("UPDATE task SET task = ? WHERE id = ? AND list_id = ?",
			task.Task, id, task.List); err != nil {
			log.Println("Failed to edit task:", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.String(403, "")
}

func deleteTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("Failed to get id param:", err)
		c.String(400, "")
		return
	}

	if checkTask(id, sessions.Default(c).Get("userID")) {
		if _, err := db.Exec("DELETE FROM task WHERE id = ?", id); err != nil {
			log.Println("Failed to delete task:", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.String(403, "")
}

func reorder(c *gin.Context) {
	var reorder struct{ List, Old, New int }
	if err := c.BindJSON(&reorder); err != nil {
		c.String(400, "")
		return
	}

	userID := sessions.Default(c).Get("userID")
	if !checkAll(reorder.Old, reorder.List, userID) || !checkAll(reorder.New, reorder.List, userID) {
		c.String(403, "")
		return
	}

	ec := make(chan error, 1)
	var oldSeq, newSeq int
	go func() {
		ec <- db.QueryRow("SELECT seq FROM seq WHERE task_id = ?",
			reorder.Old).Scan(&oldSeq)
	}()
	if err := db.QueryRow("SELECT seq FROM seq WHERE task_id = ?",
		reorder.New).Scan(&newSeq); err != nil {
		log.Println("Failed to scan new seq:", err)
		c.String(500, "")
		return
	}
	if err := <-ec; err != nil {
		log.Println("Failed to scan old seq:", err)
		c.String(500, "")
		return
	}

	var err error
	if oldSeq > newSeq {
		_, err = db.Exec("UPDATE seq SET seq = seq+1 WHERE seq >= ? AND seq < ? AND list_id = ?",
			newSeq, oldSeq, reorder.List)
	} else {
		_, err = db.Exec("UPDATE seq SET seq = seq-1 WHERE seq > ? AND seq <= ? AND list_id = ?",
			oldSeq, newSeq, reorder.List)
	}
	if err != nil {
		log.Println("Failed to update other seq:", err)
		c.String(500, "")
		return
	}
	if _, err := db.Exec("UPDATE seq SET seq = ? WHERE task_id = ?",
		newSeq, reorder.Old); err != nil {
		log.Println("Failed to update seq:", err)
		c.String(500, "")
		return
	}
	c.String(200, "1")
}
