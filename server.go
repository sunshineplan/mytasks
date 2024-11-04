package main

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb"
)

func run() error {
	if err := initDB(); err != nil {
		return err
	}

	router := gin.Default()
	router.TrustedPlatform = "X-Real-IP"
	server.Handler = router

	js, err := os.ReadFile(joinPath(dir(self), "dist/const.js"))
	if err != nil {
		return err
	}

	if *universal {
		var redisStore struct{ Endpoint, Password, Secret, API string }
		if err := meta.Get("account_redis", &redisStore); err != nil {
			return err
		}

		js = bytes.ReplaceAll(js, []byte("@universal@"), []byte(redisStore.API))

		store, err := redis.NewStore(10, "tcp", redisStore.Endpoint, redisStore.Password, []byte(redisStore.Secret))
		if err != nil {
			return err
		}
		if err := redis.SetKeyPrefix(store, "account_"); err != nil {
			return err
		}
		router.Use(sessions.Sessions("universal", store))
	} else {
		js = bytes.ReplaceAll(js, []byte("@universal@"), nil)

		secret := make([]byte, 16)
		if _, err := rand.Read(secret); err != nil {
			return err
		}
		router.Use(sessions.Sessions("session", cookie.NewStore(secret)))
	}

	if priv != nil {
		pubkey_bytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
		if err != nil {
			return err
		}
		js = bytes.ReplaceAll(
			js, []byte("@pubkey@"),
			bytes.ReplaceAll(
				pem.EncodeToMemory(&pem.Block{
					Type:  "RSA PUBLIC KEY",
					Bytes: pubkey_bytes,
				}),
				[]byte{'\n'},
				nil,
			),
		)
	} else {
		js = bytes.ReplaceAll(js, []byte("@pubkey@"), nil)
	}

	if err := os.WriteFile(joinPath(dir(self), "dist/env.js"), js, 0644); err != nil {
		return err
	}

	router.StaticFS("/assets", http.Dir(joinPath(dir(self), "dist/assets")))
	router.StaticFile("env.js", joinPath(dir(self), "dist/env.js"))
	router.StaticFile("favicon.png", joinPath(dir(self), "dist/favicon.png"))
	router.LoadHTMLFiles(joinPath(dir(self), "dist/index.html"))
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	if !*universal {
		auth := router.Group("/")
		auth.POST("/login", login)
		auth.POST("/logout", authRequired, func(c *gin.Context) {
			session := sessions.Default(c)
			session.Clear()
			session.Options(sessions.Options{MaxAge: -1})
			if err := session.Save(); err != nil {
				svc.Print(err)
				c.Status(500)
				return
			}
			c.String(200, "bye")
		})
		auth.POST("/chgpwd", authRequired, chgpwd)
	}

	router.GET("/info", func(c *gin.Context) {
		user, _ := getUser(c)
		if user.Username == "" {
			c.Status(200)
			return
		}
		c.Set("last", user.Last)
		last, ok := checkLastModified(c)
		c.SetCookie("last", last, 856400*365, "", "", false, false)
		if ok {
			c.String(200, user.Username)
		} else {
			c.Status(409)
		}
	})
	router.GET("/poll", func(c *gin.Context) {
		time.Sleep(*poll)
		user, err := getUser(c)
		if err == errNoUser || err == mongodb.ErrNoDocuments {
			c.Status(401)
		} else if err == nil {
			if v, _ := c.Cookie("last"); v == user.Last {
				c.Status(200)
			} else {
				c.String(200, user.Last)
			}
		} else {
			svc.Print(err)
			c.Status(500)
		}
	})

	base := router.Group("/")
	base.Use(authRequired, checkRequired)
	base.POST("/list/get", getList)
	base.POST("/list/edit", editList)
	base.POST("/list/delete", deleteList)
	base.POST("/task/get", getTask)
	base.POST("/task/add", addIncomplete)
	base.POST("/task/edit/:id", editIncomplete)
	base.POST("/task/complete/:id", completeTask)
	base.POST("/task/delete/:id", deleteIncomplete)
	base.POST("/task/reorder", reorder)
	base.POST("/completed/more", moreCompleted)
	base.POST("/completed/revert/:id", revertCompleted)
	base.POST("/completed/delete/:id", deleteCompleted)
	base.POST("/completed/empty", emptyCompleted)

	router.NoRoute(func(c *gin.Context) {
		c.Redirect(302, "/")
	})

	return server.Run()
}
