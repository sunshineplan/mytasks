package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb/api"
	"github.com/sunshineplan/password"
)

type user struct {
	ID       string `json:"_id"`
	Username string
	Password string
}

func authRequired(c *gin.Context) {
	if sessions.Default(c).Get("id") == nil {
		c.AbortWithStatus(401)
	}
}

func getUser(c *gin.Context) (id, username string, err error) {
	session := sessions.Default(c)
	sid := session.Get("id")
	username, _ = session.Get("username").(string)
	if universal {
		var user user
		if err = accountClient.FindOne(api.M{"uid": sid}, nil, &user); err != nil {
			return
		}
		id = user.ID
		return
	}
	id, _ = sid.(string)
	return
}

func login(c *gin.Context) {
	var login struct {
		Username, Password string
		Rememberme         bool
	}
	if err := c.BindJSON(&login); err != nil {
		c.String(400, "")
		return
	}
	login.Username = strings.ToLower(login.Username)

	if password.IsMaxAttempts(c.ClientIP() + login.Username) {
		c.JSON(200, gin.H{"status": 0, "message": fmt.Sprintf("Max retries exceeded (%d)", maxRetry)})
		return
	}

	var user user
	var message string
	if err := accountClient.FindOne(api.M{"username": login.Username}, nil, &user); err != nil {
		if err == api.ErrNoDocuments {
			message = "Incorrect username"
		} else {
			log.Print(err)
			c.String(500, "Critical Error! Please contact your system administrator.")
			return
		}
	} else {
		if priv == nil {
			_, err = password.Compare(c.ClientIP()+login.Username, user.Password, login.Password, false)
		} else {
			_, err = password.CompareRSA(c.ClientIP()+login.Username, user.Password, login.Password, false, priv)
		}
		if err != nil {
			if errors.Is(err, password.ErrIncorrectPassword) {
				message = err.Error()
			} else {
				log.Print(err)
				c.String(500, "Internal Server Error")
				return
			}
		}

		if message == "" {
			session := sessions.Default(c)
			session.Clear()
			session.Set("id", user.ID)
			session.Set("username", user.Username)

			if login.Rememberme {
				session.Options(sessions.Options{HttpOnly: true, MaxAge: 856400 * 365})
			} else {
				session.Options(sessions.Options{HttpOnly: true})
			}

			if err := session.Save(); err != nil {
				log.Print(err)
				c.String(500, "Internal Server Error")
				return
			}

			c.JSON(200, gin.H{"status": 1})
			return
		}
	}

	c.JSON(200, gin.H{"status": 0, "message": message})
}

func chgpwd(c *gin.Context) {
	session := sessions.Default(c)
	userID, username := session.Get("id"), session.Get("username")
	if userID == nil || username == nil {
		c.String(401, "")
		return
	}

	if password.IsMaxAttempts(c.ClientIP() + username.(string)) {
		c.JSON(200, gin.H{"status": 0, "message": fmt.Sprintf("Max retries exceeded (%d)", maxRetry), "error": 1})
		return
	}

	var data struct{ Password, Password1, Password2 string }
	if err := c.BindJSON(&data); err != nil {
		c.String(400, "")
		return
	}

	var user user
	if err := accountClient.FindOne(api.M{"_id": api.ObjectID(userID.(string))}, nil, &user); err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	var err error
	var message, newPassword string
	var errorCode int
	if priv == nil {
		newPassword, err = password.Change(
			c.ClientIP()+user.Username, user.Password, data.Password, data.Password1, data.Password2, false,
		)
	} else {
		newPassword, err = password.ChangeRSA(
			c.ClientIP()+user.Username, user.Password, data.Password, data.Password1, data.Password2, false, priv,
		)
	}
	if err != nil {
		message = err.Error()
		switch {
		case errors.Is(err, password.ErrIncorrectPassword):
			errorCode = 1
		case err == password.ErrConfirmPasswordNotMatch || err == password.ErrSamePassword:
			errorCode = 2
		case err == password.ErrBlankPassword:
		default:
			log.Print(err)
			c.String(500, "Internal Server Error")
			return
		}
	}

	if message == "" {
		if _, err := accountClient.UpdateOne(
			api.M{"_id": api.ObjectID(userID.(string))},
			api.M{"$set": api.M{"password": newPassword}},
			nil,
		); err != nil {
			log.Print(err)
			c.String(500, "")
			return
		}

		session.Clear()
		session.Options(sessions.Options{MaxAge: -1})
		if err := session.Save(); err != nil {
			log.Print(err)
			c.String(500, "")
			return
		}

		c.JSON(200, gin.H{"status": 1})
		return
	}

	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}
