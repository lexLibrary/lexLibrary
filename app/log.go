// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
	"github.com/rs/xid"
)

// Log is a logged error message in the database
type Log struct {
	ID       xid.ID    `json:"id"`
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

// Performance of this search will be poor, and I may just remove this functionality altogether
// but it's an admin only thing, so maybe it's worth keeping around.  There is no nice way to do case-insensitve
// columns across all of the databases, unless I drop support for TiDB, Cockroach, and DB2.  It might be
// worth it if other features require case insensitivity, but I think we can get by without it.
var sqlLogSearch = data.NewQuery(`
	select id, occurred, message from logs where lower(message) like lower({{arg "search"}}) order by occurred desc 
	{{if sqlserver}}
		OFFSET {{arg "offset"}} ROWS FETCH NEXT {{arg "limit"}} ROWS ONLY
	{{else}}
		LIMIT {{arg "limit" }} OFFSET {{arg "offset"}}
	{{end}}
`)

var sqlLogGetByID = data.NewQuery(`select id, occurred, message from logs where id = {{arg "id"}}`)

// LogError logs an error to the logs table
func LogError(lerr error) xid.ID {
	l := Log{
		ID:       xid.New(),
		Message:  lerr.Error(),
		Occurred: time.Now(),
	}

	log.Printf("ERROR: %s", l.Message)

	_, err := sqlLogInsert.Exec(
		sql.Named("id", l.ID),
		sql.Named("occurred", l.Occurred),
		sql.Named("message", l.Message))

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

	rows, err := sqlLogGet.Query(sql.Named("offset", offset), sql.Named("limit", limit))
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

// LogGetByID retrieves logs from the error log in the database for the given ID
func LogGetByID(id xid.ID) (*Log, error) {
	log := &Log{}

	err := sqlLogGetByID.QueryRow(sql.Named("id", id)).Scan(&log.ID, &log.Occurred, &log.Message)
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
		sql.Named("search", "%"+search+"%"),
		sql.Named("offset", offset),
		sql.Named("limit", limit))
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
