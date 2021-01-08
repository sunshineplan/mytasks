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
