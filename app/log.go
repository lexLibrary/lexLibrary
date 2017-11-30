// Copyright (c) 2017 Townsourced Inc.

package app

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
)

// Log is a logged error message in the database
type Log struct {
	Message  string
	Occurred time.Time
}

var sqlLogInsert = data.NewQuery(`insert into logs (occurred, message) values ({{arg "occurred"}}, {{arg "message"}})`)
var sqlLogGet = data.NewQuery(`
	select occurred, message from logs order by occurred desc 
	LIMIT {{arg "limit" }} OFFSET {{arg "offset"}}
`)
var sqlLogSearch = data.NewQuery(`
	select occurred, message from logs where message like {{arg "search"}} order by occurred desc 
	LIMIT {{arg "limit" }} OFFSET {{arg "offset"}}
`) //TODO: Make case insensitive?

// LogError logs an error to the logs table
func LogError(lerr error) {
	l := Log{
		Message:  lerr.Error(),
		Occurred: time.Now(),
	}

	log.Printf("ERROR: %s", l.Message)

	_, err := sqlLogInsert.Exec(
		sql.Named("occurred", l.Occurred),
		sql.Named("message", l.Message))

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
		err = rows.Scan(&log.Occurred, &log.Message)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// LogSearch retrieves logs from the error log in the database that contain the search value in it's message
func LogSearch(search string, offset, limit int) ([]*Log, error) {
	if limit == 0 || limit > maxRows {
		limit = 10
	}
	var logs []*Log

	rows, err := sqlLogSearch.Query(
		sql.Named("search", "%"+search+"%"),
		sql.Named("offset", offset),
		sql.Named("limit", limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		log := &Log{}
		err = rows.Scan(&log.Occurred, &log.Message)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

type logWriter struct {
}

// Write implements io.Writer by writing the bytes to the Log table
// Each call to write generates a new entry in the database
func (l *logWriter) Write(p []byte) (n int, err error) {
	LogError(fmt.Errorf("%s", p))
	return len(p), nil
}

// Logger returns a new logger instance that writes the logs to the
// database Log table.
func Logger(prefix string) *log.Logger {
	return log.New(&logWriter{}, prefix, log.Llongfile)
}
