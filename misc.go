package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sunshineplan/utils/archive"
	"github.com/sunshineplan/utils/mail"
)

func addUser(username string) {
	log.Print("Start!")
	if _, err := db.Exec("INSERT INTO user(username) VALUES (?)", strings.ToLower(username)); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			log.Fatalf("Username %s already exists.", strings.ToLower(username))
		} else {
			log.Fatalln("Failed to add user:", err)
		}
	}
	log.Print("Done!")
}

func deleteUser(username string) {
	log.Print("Start!")
	res, err := db.Exec("DELETE FROM user WHERE username=?", strings.ToLower(username))
	if err != nil {
		log.Fatalln("Failed to delete user:", err)
	}
	if n, err := res.RowsAffected(); err != nil {
		log.Fatalln("Failed to get affected rows:", err)
	} else if n == 0 {
		log.Fatalf("User %s does not exist.", strings.ToLower(username))
	}
	log.Print("Done!")
}

func backup() {
	log.Print("Start!")
	var config struct {
		SMTPServer     string
		SMTPServerPort int
		From, Password string
		To             []string
	}
	if err := meta.Get("mytasks_backup", &config); err != nil {
		log.Fatalln("Failed to get mytasks_backup metadata:", err)
	}
	dialer := &mail.Dialer{
		Host:     config.SMTPServer,
		Port:     config.SMTPServerPort,
		Account:  config.From,
		Password: config.Password,
	}

	tmpfile, err := ioutil.TempFile("", "tmp")
	if err != nil {
		log.Fatalln("Failed to create temporary file:", err)
	}
	tmpfile.Close()
	if err := dbConfig.Backup(tmpfile.Name()); err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	b, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	if err := archive.Pack(&buf, archive.ZIP, archive.File{Name: "database", Body: b}); err != nil {
		log.Fatal(err)
	}
	if err := dialer.Send(
		&mail.Message{
			To:          config.To,
			Subject:     fmt.Sprintf("My Tasks Backup-%s", time.Now().Format("20060102")),
			Attachments: []*mail.Attachment{{Bytes: buf.Bytes(), Filename: "backup.zip"}},
		},
	); err != nil {
		log.Fatalln("Failed to send mail:", err)
	}
	log.Print("Done!")
}

func restore(file string) {
	log.Print("Start!")
	if file == "" {
		file = joinPath(dir(self), "scripts/schema.sql")
	} else {
		if _, err := os.Stat(file); err != nil {
			log.Fatalln("File not found:", err)
		}
	}
	dropAll := joinPath(dir(self), "scripts/drop_all.sql")
	if err := dbConfig.Restore(dropAll); err != nil {
		log.Fatal(err)
	}
	if err := dbConfig.Restore(file); err != nil {
		log.Fatal(err)
	}
	log.Print("Done!")
}
