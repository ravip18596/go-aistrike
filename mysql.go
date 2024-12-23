package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"time"
)

func InitMysql() *sql.DB {
	db, err := sql.Open("mysql", "aistrike:aistrike@tcp(mysql:3306)/aistrike")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":    err.Error(),
			"database": "aistrike",
		}).Error("Error connecting to database aistrike")
		return nil
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return db
}

var mysql = InitMysql()
