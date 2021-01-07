package main

import (
	"database/sql"
	"time"

	"github.com/sunshineplan/utils/database/mysql"
)

var dbConfig mysql.Config
var db *sql.DB

func initDB() (err error) {
	//if err = meta.Get("mytasks_mysql", &dbConfig); err != nil {
	//	return
	//}
	dbConfig.Server = "localhost"
	dbConfig.Port = 3306
	dbConfig.Database = "mytasks"
	dbConfig.Username = "mytasks"
	dbConfig.Password = "123456ab"

	db, err = dbConfig.Open()
	if err != nil {
		return err
	}
	db.SetConnMaxLifetime(time.Minute * 1)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return nil
}
