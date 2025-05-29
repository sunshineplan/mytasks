package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb"
	"github.com/sunshineplan/password"
	"github.com/sunshineplan/utils/cache"
)

var userCache = cache.NewWithRenew[string, user](true)

type user struct {
	ID       mongodb.OID `json:"_id" bson:"_id"`
	Username string
	Password string
	Last     string
}

func authRequired(c *gin.Context) {
	if user, err := getUser(c); err == errNoUser || err == mongodb.ErrNoDocuments {
		c.AbortWithStatus(401)
	} else if err == nil {
		c.Set("id", user.ID.Hex())
		c.Set("username", user.Username)
		c.Set("last", user.Last)
	} else {
		svc.Print(err)
		c.AbortWithStatus(500)
	}
}

var errNoUser = errors.New("no user")

func getUser(c *gin.Context) (usr user, err error) {
	id := sessions.Default(c).Get("id")
	if id == nil {
		err = errNoUser
		return
	}
	var ok bool
	if usr, ok = userCache.Get(id.(string)); ok {
		return
	}
	var filter mongodb.M
	if *universal {
		filter = mongodb.M{"uid": id}
	} else {
		id, _ := accountClient.ObjectID(id.(string))
		filter = mongodb.M{"_id": id}
	}
	mu.Lock()
	defer mu.Unlock()
	if err = accountClient.FindOne(filter, nil, &usr); err != nil {
		return
	}
	userCache.Set(usr.ID.Hex(), usr, 24*time.Hour, nil)
	return
}

type auth struct {
	username string
	ip       string
}

func login(c *gin.Context) {
	var login struct {
		Username, Password string
		Rememberme         bool
	}
	if err := c.BindJSON(&login); err != nil {
		c.Status(400)
		return
	}
	login.Username = strings.ToLower(login.Username)

	if password.IsMaxAttempts(auth{login.Username, c.ClientIP()}) {
		c.JSON(200, gin.H{"status": 0, "message": fmt.Sprintf("Max retries exceeded (%d)", *maxRetry)})
		return
	}

	var user user
	var message string
	if err := accountClient.FindOne(mongodb.M{"username": login.Username}, nil, &user); err != nil {
		if err == mongodb.ErrNoDocuments {
			message = "Incorrect username"
		} else {
			svc.Print(err)
			c.String(500, "Critical Error! Please contact your system administrator.")
			return
		}
	} else {
		if err = password.CompareHashAndPassword(auth{login.Username, c.ClientIP()}, user.Password, login.Password); err != nil {
			if errors.Is(err, password.ErrIncorrectPassword) {
				message = err.Error()
			} else if user.Password != login.Password {
				svc.Print(err)
				c.String(500, "Internal Server Error")
				return
			}
		}

		if message == "" {
			session := sessions.Default(c)
			session.Clear()
			session.Set("id", user.ID.Hex())

			if login.Rememberme {
				session.Options(sessions.Options{HttpOnly: true, MaxAge: 856400 * 365})
			} else {
				session.Options(sessions.Options{HttpOnly: true})
			}

			if err := session.Save(); err != nil {
				svc.Print(err)
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

	if password.IsMaxAttempts(auth{username.(string), c.ClientIP()}) {
		c.JSON(200, gin.H{"status": 0, "message": fmt.Sprintf("Max retries exceeded (%d)", *maxRetry), "error": 1})
		return
	}

	var data struct{ Password, Password1, Password2 string }
	if err := c.BindJSON(&data); err != nil {
		c.Status(400)
		return
	}
	var err error
	if priv != nil {
		data.Password1, err = password.DecryptPKCS1v15(priv, data.Password1)
		if err != nil {
			c.String(400, "Bad Request")
			return
		}
		data.Password2, err = password.DecryptPKCS1v15(priv, data.Password2)
		if err != nil {
			c.String(400, "Bad Request")
			return
		}
	}

	id, _ := accountClient.ObjectID(userID.(string))
	var user user
	if err := accountClient.FindOne(mongodb.M{"_id": id}, nil, &user); err != nil {
		svc.Print(err)
		c.Status(500)
		return
	}

	var message string
	var errorCode int
	if err = password.CompareHashAndPassword(auth{username.(string), c.ClientIP()}, user.Password, data.Password); err != nil {
		if errors.Is(err, password.ErrIncorrectPassword) {
			message = err.Error()
			errorCode = 1
		} else {
			svc.Print(err)
			c.String(500, "Internal Server Error")
			return
		}
	} else {
		if priv != nil {
			data.Password, _ = password.DecryptPKCS1v15(priv, data.Password)
		}
		switch {
		case data.Password1 != data.Password2:
			message = "confirm password doesn't match new password"
			errorCode = 2
		case data.Password1 == data.Password:
			message = "new password cannot be the same as old password"
			errorCode = 2
		case data.Password1 == "":
			message = "new password cannot be blank"
		}
	}

	if message == "" {
		newPassword, err := password.HashPassword(data.Password1)
		if err != nil {
			svc.Print(err)
			c.String(500, "Internal Server Error")
			return
		}
		if _, err := accountClient.UpdateOne(
			mongodb.M{"_id": id},
			mongodb.M{"$set": mongodb.M{"password": newPassword}},
			nil,
		); err != nil {
			svc.Print(err)
			c.Status(500)
			return
		}

		session.Clear()
		session.Options(sessions.Options{MaxAge: -1})
		if err := session.Save(); err != nil {
			svc.Print(err)
			c.Status(500)
			return
		}

		c.JSON(200, gin.H{"status": 1})
		return
	}

	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}
