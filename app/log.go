// Copyright (c) 2017 Townsourced Inc.

package app

import (
	"database/sql"
	"log"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
)

// Log is a logged error message in the database
type Log struct {
	message  string
	occurred time.Time
}

var sqlLogInsert = data.NewQuery(`insert into logs (occurred, message) values ({{arg "occurred"}}, {{arg "message"}})`)
var sqlLogGet = data.NewQuery(`
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

	_, err := sqlLogInsert.Exec(
		sql.Named("occurred", l.occurred),
		sql.Named("message", l.message))

	if err != nil {
		log.Printf(`Error inserting error log entry. Log entry: %s ERROR: %s`, lerr, err)
	}
}

// LogGet retrieves logs from the error log in the database
func LogGet(offset, limit int) ([]*Log, error) {
	if limit == 0 || limit > maxRows {
		limit = 10
	}
	var logs []*Log

	rows, err := sqlLogGet.Query(sql.Named("offset", offset), sql.Named("limit", limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		log := &Log{}
		err = rows.Scan(&log.occurred, &log.message)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}
