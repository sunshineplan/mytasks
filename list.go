package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type list struct {
	ID         int    `json:"id"`
	Name       string `json:"list"`
	Incomplete int    `json:"incomplete"`
	Completed  int    `json:"completed"`
}

func checkList(listID, userID interface{}) bool {
	var exist string
	if err := db.QueryRow("SELECT list FROM list WHERE id = ? AND user_id = ?",
		listID, userID).Scan(&exist); err == nil {
		return true
	}
	return false
}

func getList(userID interface{}) ([]list, error) {
	rows, err := db.Query("SELECT id, list, incomplete, completed FROM lists WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lists := []list{}
	for rows.Next() {
		var list list
		if err := rows.Scan(&list.ID, &list.Name, &list.Incomplete, &list.Completed); err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}

	return lists, nil
}

func addList(c *gin.Context) {
	userID := sessions.Default(c).Get("userID")

	var list list
	if err := c.BindJSON(&list); err != nil {
		c.String(400, "")
		return
	}
	list.Name = strings.TrimSpace(list.Name)

	var message string
	switch {
	case list.Name == "":
		message = "List name is empty."
	case len(list.Name) > 15:
		message = "List name exceeded length limit."
	default:
		var exist string
		if err := db.QueryRow("SELECT id FROM list WHERE list = ? AND user_id = ?",
			list.Name, userID).Scan(&exist); err == nil {
			message = fmt.Sprintf("List %s is already existed.", list.Name)
		} else {
			if err == sql.ErrNoRows {
				result, err := db.Exec("INSERT INTO list (list, user_id) VALUES (?, ?)",
					list.Name, userID)
				if err != nil {
					log.Println("Failed to add list:", err)
					c.String(500, "")
					return
				}
				id, err := result.LastInsertId()
				if err != nil {
					log.Println("Failed to get last insert id:", err)
					c.String(500, "")
					return
				}
				c.JSON(200, gin.H{"status": 1, "id": id})
				return
			}
			log.Println("Failed to scan list:", err)
			c.String(500, "")
			return
		}
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": 1})
}

func editList(c *gin.Context) {
	userID := sessions.Default(c).Get("userID")

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("Failed to get id param:", err)
		c.String(400, "")
		return
	}

	var list list
	if err := c.BindJSON(&list); err != nil {
		c.String(400, "")
		return
	}
	list.Name = strings.TrimSpace(list.Name)

	var exist string
	err = db.QueryRow("SELECT id FROM list WHERE list = ? AND user_id = ? AND id != ?",
		list.Name, userID, id).Scan(&exist)

	var message string
	switch {
	case list.Name == "":
		message = "New list name is empty."
	case len(list.Name) > 15:
		message = "List name exceeded length limit."
	case err == nil:
		message = fmt.Sprintf("List %s is already existed.", list.Name)
	default:
		if _, err := db.Exec("UPDATE list SET list = ? WHERE id = ? AND user_id = ?",
			list.Name, id, userID); err != nil {
			log.Println("Failed to edit list:", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message})
}

func deleteList(c *gin.Context) {
	userID := sessions.Default(c).Get("userID")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("Failed to get id param:", err)
		c.String(400, "")
		return
	}

	if _, err := db.Exec("DELETE FROM list WHERE id = ? and user_id = ?",
		id, userID); err != nil {
		log.Println("Failed to delete list:", err)
		c.String(500, "")
		return
	}
	if _, err := db.Exec("DELETE FROM task WHERE list_id = ?", id); err != nil {
		log.Println("Failed to remove deleted list for task:", err)
		c.String(500, "")
		return
	}
	c.JSON(200, gin.H{"status": 1})
}

func reorderList(c *gin.Context) {
	var reorder struct{ Old, New int }
	if err := c.BindJSON(&reorder); err != nil {
		c.String(400, "")
		return
	}

	userID := sessions.Default(c).Get("userID")
	if !checkList(reorder.Old, userID) || !checkList(reorder.New, userID) {
		c.String(403, "")
		return
	}

	ec := make(chan error, 1)
	var oldSeq, newSeq int
	go func() {
		ec <- db.QueryRow("SELECT seq FROM list_seq WHERE list_id = ?",
			reorder.Old).Scan(&oldSeq)
	}()
	if err := db.QueryRow("SELECT seq FROM list_seq WHERE list_id = ?",
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
		_, err = db.Exec("UPDATE list_seq SET seq = seq+1 WHERE seq >= ? AND seq < ? AND user_id = ?",
			newSeq, oldSeq, userID)
	} else {
		_, err = db.Exec("UPDATE list_seq SET seq = seq-1 WHERE seq > ? AND seq <= ? AND user_id = ?",
			oldSeq, newSeq, userID)
	}
	if err != nil {
		log.Println("Failed to update other seq:", err)
		c.String(500, "")
		return
	}
	if _, err := db.Exec("UPDATE list_seq SET seq = ? WHERE list_id = ?",
		newSeq, reorder.Old); err != nil {
		log.Println("Failed to update seq:", err)
		c.String(500, "")
		return
	}
	c.String(200, "1")
}
