package main

import (
	"crypto/rand"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func run() {
	if logPath != "" {
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			log.Fatalln("Failed to open log file:", err)
		}
		gin.DefaultWriter = f
		gin.DefaultErrorWriter = f
		log.SetOutput(f)
	}

	secret := make([]byte, 16)
	if _, err := rand.Read(secret); err != nil {
		log.Fatalln("Failed to get secret:", err)
	}

	router := gin.Default()
	server.Handler = router
	router.Use(sessions.Sessions("session", sessions.NewCookieStore(secret)))
	router.StaticFS("/build", http.Dir(joinPath(dir(self), "public/build")))
	router.StaticFile("favicon.png", joinPath(dir(self), "public/favicon.png"))
	router.StaticFile("style.css", joinPath(dir(self), "public/style.css"))
	router.LoadHTMLFiles(joinPath(dir(self), "public/index.html"))
	router.GET("/", func(c *gin.Context) {
		userID := sessions.Default(c).Get("user_id")
		if userID != nil && userID != 0 {
			if _, err := c.Cookie("Username"); err != nil {
				username, err := getUser(c)
				if err != nil {
					c.String(500, "")
					return
				}
				c.SetCookie("Username", username, 0, "", "", false, false)
			}
		} else {
			c.SetCookie("Username", "", -1, "", "", false, false)
		}
		c.HTML(200, "index.html", nil)
	})

	auth := router.Group("/")
	auth.POST("/login", login)
	auth.GET("/logout", authRequired, func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		c.SetCookie("Username", "", -1, "", "", false, false)
		session.Save()
		c.Redirect(302, "/")
	})
	auth.POST("/setting", authRequired, setting)

	base := router.Group("/")
	base.Use(authRequired)
	base.POST("/task/get", getTask)
	base.POST("/task/add", addTask)
	base.POST("/task/edit/:id", editTask)
	base.POST("/task/delete/:id", deleteTask)
	base.POST("/list/get", getList)
	base.POST("/list/add", addList)
	base.POST("/list/edit/:id", editList)
	base.POST("/list/delete/:id", deleteList)
	base.POST("/reorder", reorder)

	router.NoRoute(func(c *gin.Context) {
		c.Redirect(302, "/")
	})

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
