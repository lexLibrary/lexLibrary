// Copyright (c) 2017 Townsourced Inc.
package data

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	sqlite      = iota //https://github.com/mattn/go-sqlite3
	postgres           //https://github.com/lib/pq
	mysql              //https://github.com/go-sql-driver/mysql/
	cockroachdb        //https://github.com/lib/pq
	tidb               //https://github.com/go-sql-driver/mysql/
)

// multiQuery is query that contains multiple definitions for different database backend
// for use with databases where a query can't be shared across all DB backends
/*
	q := multiQuery{
		sqlite:   "select * from tbl LIMIT 10 OFFSET 50",
		postgres: "select * from tbl OFFSET 10 ROWS FETCH NEXT 50 ROWS ONLY",
	}

	fmt.Println(q[dbType])
*/
type multiQuery map[int]string

var db *sql.DB
var dbType int

// Config is the data layer configuration used to determine how to initialize the data layer
type Config struct {
	DatabaseFile string
	SearchFile   string

	DBType string
	URL    string

	SSLCert     string
	SSLKey      string
	SSLRootCert string

	MaxIdleConnections    int
	MaxOpenConnections    int
	MaxConnectionLifetime time.Duration
}

// DefaultConfig returns the default configuration for the data layer
func DefaultConfig() Config {
	return Config{
		DatabaseFile: "./lexLibrary.db",
		SearchFile:   "./lexLibrary.search",
	}
}

func Init(cfg Config) error {
	err := initSearch(cfg)
	if err != nil {
		return err
	}

	switch strings.ToLower(cfg.DBType) {
	case "postgres":
		err = initPostgres(cfg)
		if err != nil {
			return err
		}
	default:
		if cfg.DatabaseFile == "" {
			cfg.DatabaseFile = DefaultConfig().DatabaseFile
		}
		err = initSQLite(cfg.DatabaseFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func initSQLite(filename string) error {
	dbType = sqlite
	return fmt.Errorf("TODO")
}

func initPostgres(cfg Config) error {
	return fmt.Errorf("TODO")
}
