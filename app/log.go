// Copyright (c) 2017 Townsourced Inc.

package app

import (
	"log"
	"time"
)

// Log is a logged error message in the database
type Log struct {
	message  string
	occurred time.Time
}

// LogError logs an error to the logs table
func LogError(lerr error) {
	l := Log{
		message:  lerr.Error(),
		occurred: time.Now(),
	}

	log.Printf(l.message)

	err := l.insert()
	if err != nil {
		log.Printf(`Error inserting error log entry. Log entry: %s ERROR: %s`, lerr, err)
	}
}

// LogGet retrieves logs from the error log in the database
// func LogGet(offset, limit int) ([]Log, error) {
// 	// db.Query(queryTemplate(`select occurred, message from logs order by occurred desc {{offsetLimit offset, limit}} `), offset, limit)
// }

func (l *Log) insert() error {
	_, err := db.Exec(queryTemplate("insert into logs (occurred, message) values ({{param}}, {{param}})"), l.occurred, l.message)
	return err
}
