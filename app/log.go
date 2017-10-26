// Copyright (c) 2017 Townsourced Inc.

package app

import (
	"database/sql"
	"log"
	"time"
)

// Log is a logged error message in the database
type Log struct {
	message  string
	occurred time.Time
}

var sqlLogInsert = newQuery(`insert into logs (occurred, message) values ({{arg "occurred"}}, {{arg "message"}})`)
var sqlLogGet = newQuery(`
	select occurred, message from logs order by occurred desc 
	{{if sqlite}}
		LIMIT {{arg "limit" }} OFFSET {{arg "offset"}}
	{{else}}
		OFFSET {{arg "offset"}} ROWS FETCH NEXT {{arg "limit"}} ROWS ONLY
	{{end}}
`)

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
// 	db.Query(sqlLogGet.query(sql.Named("offset", offset), sql.Named("limit", limit)))
// }

func (l *Log) insert() error {
	_, err := sqlLogInsert.exec(
		sql.Named("occurred", l.occurred),
		sql.Named("message", l.message))
	return err
}
