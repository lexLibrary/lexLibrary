// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
)

// Log is a logged error message in the database
type Log struct {
	ID       data.ID   `json:"id"`
	Message  string    `json:"message,omitempty"`
	Occurred time.Time `json:"occurred,omitempty"`
}

var sqlLogInsert = data.NewQuery(`insert into logs (id, occurred, message) 
	values ({{arg "id"}}, {{arg "occurred"}}, {{arg "message"}})`)
var sqlLogGet = data.NewQuery(`
	select id, occurred, message from logs order by occurred desc 
	{{if sqlserver}}
		OFFSET {{arg "offset"}} ROWS FETCH NEXT {{arg "limit"}} ROWS ONLY
	{{else}}
		LIMIT {{arg "limit" }} OFFSET {{arg "offset"}}
	{{end}}
`)

var sqlLogTotal = data.NewQuery(`select count(*) from logs`)
var sqlLogTotalSince = data.NewQuery(`select count(*) from logs where occurred >= {{arg "occurred"}}`)

// Performance of this search will be poor, and I may just remove this functionality altogether
// but it's an admin only thing, so maybe it's worth keeping around.  There is no nice way to do case-insensitve
// columns across all of the databases, unless I drop support for TiDB, Cockroach.  It might be
// worth it if other features require case insensitivity, but I think we can get by without it.
var sqlLogSearch = data.NewQuery(`
	select id, occurred, message from logs 
	where lower(message) like lower({{arg "search"}}) order by occurred desc 
	{{if sqlserver}}
		OFFSET {{arg "offset"}} ROWS FETCH NEXT {{arg "limit"}} ROWS ONLY
	{{else}}
		LIMIT {{arg "limit" }} OFFSET {{arg "offset"}}
	{{end}}
`)

var sqlLogSearchTotal = data.NewQuery(`select count(*) from logs where lower(message) like lower({{arg "search"}})`)
var sqlLogGetByID = data.NewQuery(`select id, occurred, message from logs where id = {{arg "id"}}`)

// LogError logs an error to the logs table
func LogError(lerr error) data.ID {
	l := Log{
		ID:       data.NewID(),
		Message:  lerr.Error(),
		Occurred: time.Now(),
	}

	if os.Getenv("LLTEST") != "true" {
		// don't print out error logs during testing, it keeps console cleaner
		log.Printf("ERROR: %s", l.Message)
	}

	_, err := sqlLogInsert.Exec(
		data.Arg("id", l.ID),
		data.Arg("occurred", l.Occurred),
		data.Arg("message", l.Message))

	if err != nil {
		log.Printf(`Error inserting error log entry %s. Log entry: %s ERROR: %s`, l.ID, lerr, err)
	}

	return l.ID
}

// LogGet retrieves logs from the error log in the database
func LogGet(offset, limit int) ([]*Log, error) {
	if limit == 0 || limit > maxRows {
		limit = 10
	}

	var logs []*Log

	rows, err := sqlLogGet.Query(data.Arg("offset", offset), data.Arg("limit", limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		log := &Log{}
		err = rows.Scan(&log.ID, &log.Occurred, &log.Message)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

// LogTotal returns the total number of logs in the database
func LogTotal() (int, error) {
	total := 0
	err := sqlLogTotal.QueryRow().Scan(&total)
	return total, err
}

// LogTotalSince returns the number of logs since the passed in date
func LogTotalSince(date time.Time) (int, error) {
	total := 0
	err := sqlLogTotalSince.QueryRow(data.Arg("occurred", date)).Scan(&total)
	return total, err
}

// LogGetByID retrieves logs from the error log in the database for the given ID
func LogGetByID(id data.ID) (*Log, error) {
	log := &Log{}

	err := sqlLogGetByID.QueryRow(data.Arg("id", id)).Scan(&log.ID, &log.Occurred, &log.Message)
	if err != nil {
		return nil, err
	}

	return log, nil
}

// LogSearch retrieves logs from the error log in the database that contain the search value in it's message
func LogSearch(search string, offset, limit int) ([]*Log, error) {
	if limit == 0 || limit > maxRows {
		limit = 10
	}
	var logs []*Log

	rows, err := sqlLogSearch.Query(
		data.Arg("search", "%"+search+"%"),
		data.Arg("offset", offset),
		data.Arg("limit", limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		log := &Log{}
		err = rows.Scan(&log.ID, &log.Occurred, &log.Message)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// LogSearchTotal returns the total number of logs that match the given search value
func LogSearchTotal(search string) (int, error) {
	total := 0
	err := sqlLogSearchTotal.QueryRow(data.Arg("search", "%"+search+"%")).Scan(&total)
	return total, err
}

type logWriter struct{}

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
