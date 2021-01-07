package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type list struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func getListID(list string, userID int) (int, error) {
	if list != "" {
		var listID int
		err := db.QueryRow("SELECT id FROM list WHERE list = ? AND user_id = ?", list, userID).Scan(&listID)
		switch {
		case len(list) > 15:
			return -1, nil
		case err != nil:
			if err == sql.ErrNoRows {
				res, err := db.Exec("INSERT INTO list (list, user_id) VALUES (?, ?)", list, userID)
				if err != nil {
					log.Println("Failed to add list:", err)
					return 0, err
				}
				lastInsertID, err := res.LastInsertId()
				if err != nil {
					log.Println("Failed to get last insert id:", err)
					return 0, err
				}
				return int(lastInsertID), nil
			}
			return 0, err
		default:
			return listID, nil
		}
	} else {
		return 0, nil
	}
}

func getList(c *gin.Context) {
	userID := sessions.Default(c).Get("user_id")

	rows, err := db.Query("SELECT id, list, count FROM lists WHERE user_id = ?", userID)
	if err != nil {
		log.Println("Failed to get lists:", err)
		c.String(500, "")
		return
	}
	defer rows.Close()
	lists := []list{}
	for rows.Next() {
		var list list
		if err := rows.Scan(&list.ID, &list.Name, &list.Count); err != nil {
			log.Println("Failed to scan list:", err)
			c.String(500, "")
			return
		}
		lists = append(lists, list)
	}

	c.JSON(200, lists)
}

func addList(c *gin.Context) {
	userID := sessions.Default(c).Get("user_id")

	var list list
	if err := c.BindJSON(&list); err != nil {
		c.String(400, "")
		return
	}

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
				if _, err := db.Exec("INSERT INTO list (list, user_id) VALUES (?, ?)",
					list.Name, userID); err != nil {
					log.Println("Failed to add list:", err)
					c.String(500, "")
					return
				}
				c.JSON(200, gin.H{"status": 1})
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
	userID := sessions.Default(c).Get("user_id")

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

	ec := make(chan error, 1)
	var exist string
	go func() {
		ec <- db.QueryRow("SELECT id FROM list WHERE list = ? AND user_id = ?",
			list.Name, userID).Scan(&exist)
	}()
	var oldList string
	if err := db.QueryRow("SELECT list FROM list WHERE id = ? AND user_id = ?",
		id, userID).Scan(&oldList); err != nil {
		log.Println("Failed to scan list:", err)
		c.String(500, "")
		return
	}
	err = <-ec

	var message string
	var errorCode int
	switch {
	case list.Name == "":
		message = "New list name is empty."
		errorCode = 1
	case oldList == list.Name:
		message = "New list is same as old list."
	case len(list.Name) > 15:
		message = "List name exceeded length limit."
		errorCode = 1
	case err == nil:
		message = fmt.Sprintf("List %s is already existed.", list.Name)
		errorCode = 1
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
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func deleteList(c *gin.Context) {
	userID := sessions.Default(c).Get("user_id")
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
	if _, err := db.Exec("UPDATE task SET list_id = 0 WHERE list_id = ? and user_id = ?",
		id, userID); err != nil {
		log.Println("Failed to remove deleted list for task:", err)
		c.String(500, "")
		return
	}
	c.JSON(200, gin.H{"status": 1})
}
