package main

import (
	"crypto/rand"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
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

	if err := initDB(); err != nil {
		log.Fatalln("Failed to initialize database:", err)
	}

	router := gin.Default()
	server.Handler = router

	js, err := ioutil.ReadFile(joinPath(dir(self), "public/build/bundle.js"))
	if err != nil {
		log.Fatal(err)
	}

	if universal {
		var redisStore struct{ Endpoint, Password, Secret, API string }
		if err := meta.Get("account_redis", &redisStore); err != nil {
			log.Fatal(err)
		}

		if err := ioutil.WriteFile(joinPath(dir(self), "public/build/script.js"),
			[]byte(strings.ReplaceAll(string(js), "@universal@", redisStore.API)), 0644); err != nil {
			log.Fatal(err)
		}

		store, err := redis.NewStore(10, "tcp", redisStore.Endpoint, redisStore.Password, []byte(redisStore.Secret))
		if err != nil {
			log.Fatal(err)
		}
		if err := redis.SetKeyPrefix(store, "account_"); err != nil {
			log.Fatal(err)
		}
		router.Use(sessions.Sessions("universal", store))
	} else {
		if err := ioutil.WriteFile(joinPath(dir(self), "public/build/script.js"),
			[]byte(strings.ReplaceAll(string(js), "@universal@", "")), 0644); err != nil {
			log.Fatal(err)
		}

		secret := make([]byte, 16)
		if _, err := rand.Read(secret); err != nil {
			log.Fatalln("Failed to get secret:", err)
		}
		router.Use(sessions.Sessions("session", cookie.NewStore(secret)))
	}

	router.StaticFS("/build", http.Dir(joinPath(dir(self), "public/build")))
	router.StaticFile("favicon.png", joinPath(dir(self), "public/favicon.png"))
	router.StaticFile("style.css", joinPath(dir(self), "public/style.css"))
	router.LoadHTMLFiles(joinPath(dir(self), "public/index.html"))
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	if !universal {
		auth := router.Group("/")
		auth.POST("/login", login)
		auth.POST("/logout", authRequired, func(c *gin.Context) {
			session := sessions.Default(c)
			session.Clear()
			session.Options(sessions.Options{MaxAge: -1})
			if err := session.Save(); err != nil {
				log.Print(err)
				c.String(500, "")
				return
			}
			c.String(200, "bye")
		})
		auth.POST("/chgpwd", authRequired, chgpwd)
	}

	router.GET("/info", info)

	base := router.Group("/")
	base.Use(authRequired)
	base.POST("/get", get)
	base.POST("/list/add", addList)
	base.POST("/list/edit/:id", editList)
	base.POST("/list/delete/:id", deleteList)
	base.POST("/list/reorder", reorderList)
	base.POST("/task/add", addTask)
	base.POST("/task/edit/:id", editTask)
	base.POST("/task/complete/:id", completeTask)
	base.POST("/task/delete/:id", deleteTask)
	base.POST("/task/reorder", reorderTask)
	base.POST("/completed/more", moreCompleted)
	base.POST("/completed/revert/:id", revertCompleted)
	base.POST("/completed/delete/:id", deleteCompleted)
	base.POST("/completed/empty/:id", emptyCompleted)

	router.NoRoute(func(c *gin.Context) {
		c.Redirect(302, "/")
	})

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
