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
	ID    int    `json:"id"`
	Name  string `json:"list"`
	Count int    `json:"count"`
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
	rows, err := db.Query("SELECT id, list, count FROM lists WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lists := []list{}
	for rows.Next() {
		var list list
		if err := rows.Scan(&list.ID, &list.Name, &list.Count); err != nil {
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
