package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type task struct {
	ID   int    `json:"id"`
	Task string `json:"task"`
	List string `json:"list"`
}

func getTask(c *gin.Context) {
	var r struct{ List int }
	if err := c.BindJSON(&r); err != nil {
		c.String(400, "")
		return
	}

	userID := sessions.Default(c).Get("userID")

	stmt := "SELECT %s FROM tasks WHERE"

	var args []interface{}
	switch r.List {
	case -1:
		stmt += " user_id = ?"
		args = append(args, userID)
	case 0:
		stmt += " list_id = 0 AND user_id = ?"
		args = append(args, userID)
	default:
		stmt += " list_id = ? AND user_id = ?"
		args = append(args, r.List)
		args = append(args, userID)
	}

	rows, err := db.Query(fmt.Sprintf(stmt, "task_id, task, list"), args...)
	if err != nil {
		log.Println("Failed to get tasks:", err)
		c.String(500, "")
		return
	}
	defer rows.Close()
	tasks := []task{}
	for rows.Next() {
		var task task
		var listByte []byte
		if err := rows.Scan(&task.ID, &task.Task, &listByte); err != nil {
			log.Println("Failed to scan tasks:", err)
			c.String(500, "")
			return
		}
		task.List = string(listByte)
		tasks = append(tasks, task)
	}
	c.JSON(200, tasks)
}

func addTask(c *gin.Context) {
	userID := sessions.Default(c).Get("userID")

	var task task
	if err := c.BindJSON(&task); err != nil {
		c.String(400, "")
		return
	}

	listID, err := getListID(task.List, userID.(int))
	if err != nil {
		log.Println("Failed to get list id:", err)
		c.String(500, "")
		return
	}

	var message string
	var errorCode int
	switch {
	case task.Task == "":
		message = "Task is empty."
		errorCode = 1
	case listID == -1:
		message = "List name exceeded length limit."
		errorCode = 2
	default:
		if _, err := db.Exec("INSERT INTO task (task, user_id, list_id) VALUES (?, ?, ?)",
			task.Task, userID, listID); err != nil {
			log.Println("Failed to add task:", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func editTask(c *gin.Context) {
	userID := sessions.Default(c).Get("userID")

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("Failed to get id param:", err)
		c.String(400, "")
		return
	}
	var new task
	if err := c.BindJSON(&new); err != nil {
		c.String(400, "")
		return
	}

	bc := make(chan error, 1)
	var old task
	go func() {
		var oldList []byte
		bc <- db.QueryRow("SELECT task, list FROM tasks WHERE task_id = ? AND user_id = ?",
			id, userID).Scan(&old.Task, &oldList)
		old.List = string(oldList)
	}()
	listID, err := getListID(new.List, userID.(int))
	if err != nil {
		log.Println("Failed to get list id:", err)
		c.String(500, "")
		return
	}

	if err := <-bc; err != nil {
		log.Println(err)
		c.String(500, "")
		return
	}

	var message string
	var errorCode int
	switch {
	case new.Task == "":
		message = "Task is empty."
		errorCode = 1
	case old == new:
		message = "New task is same as old task."
	case listID == -1:
		message = "List name exceeded length limit."
		errorCode = 2
	default:
		if _, err := db.Exec("UPDATE task SET task = ?, list_id = ? WHERE id = ? AND user_id = ?",
			new.Task, listID, id, userID); err != nil {
			log.Println("Failed to edit task:", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func deleteTask(c *gin.Context) {
	userID := sessions.Default(c).Get("userID")

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("Failed to get id param:", err)
		c.String(400, "")
		return
	}

	if _, err := db.Exec("DELETE FROM task WHERE id = ? and user_id = ?", id, userID); err != nil {
		log.Println("Failed to delete task:", err)
		c.String(500, "")
		return
	}
	c.JSON(200, gin.H{"status": 1})
}

func reorder(c *gin.Context) {
	userID := sessions.Default(c).Get("userID")

	var reorder struct{ Old, New int }
	if err := c.BindJSON(&reorder); err != nil {
		c.String(400, "")
		return
	}

	ec := make(chan error, 1)
	var oldSeq, newSeq int
	go func() {
		ec <- db.QueryRow("SELECT seq FROM seq WHERE task_id = ? AND user_id = ?",
			reorder.Old, userID).Scan(&oldSeq)
	}()
	if err := db.QueryRow("SELECT seq FROM seq WHERE task_id = ? AND user_id = ?",
		reorder.New, userID).Scan(&newSeq); err != nil {
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
		_, err = db.Exec("UPDATE seq SET seq = seq+1 WHERE seq >= ? AND seq < ? AND user_id = ?",
			newSeq, oldSeq, userID)
	} else {
		_, err = db.Exec("UPDATE seq SET seq = seq-1 WHERE seq > ? AND seq <= ? AND user_id = ?",
			oldSeq, newSeq, userID)
	}
	if err != nil {
		log.Println("Failed to update other seq:", err)
		c.String(500, "")
		return
	}
	if _, err := db.Exec("UPDATE seq SET seq = ? WHERE task_id = ? AND user_id = ?",
		newSeq, reorder.Old, userID); err != nil {
		log.Println("Failed to update seq:", err)
		c.String(500, "")
		return
	}
	c.String(200, "1")
}
