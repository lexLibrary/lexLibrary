// Copyright (c) 2017 Townsourced Inc.

// Package app handles all the data and application structures handling for Lex Library
package app

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // register sqlite3
)

// Database Types
const (
	sqlite      = iota // github.com/mattn/go-sqlite3
	postgres           // github.com/lib/pq
	mysql              // github.com/go-sql-driver/mysql/
	cockroachdb        // github.com/lib/pq
	tidb               // github.com/go-sql-driver/mysql/
)

const databaseName = "lexLibrary"

const maxRows = 10000

var db *sql.DB
var dbType int

// Config is the data layer configuration used to determine how to initialize the data layer
type Config struct {
	DatabaseFile string
	SearchFile   string

	DatabaseType string
	DatabaseURL  string

	SSLCert     string
	SSLKey      string
	SSLRootCert string

	MaxIdleConnections    int
	MaxOpenConnections    int
	MaxConnectionLifetime string

	AllowSchemaRollback bool
}

// DefaultConfig returns the default configuration for the data layer
func DefaultConfig() Config {
	return Config{
		DatabaseFile: "./lexLibrary.db",
		SearchFile:   "./lexLibrary.search",
	}
}

// Init initializes the data layer based on the passed in configuration.
// Initialization includes things like setting up the database and the connections to it.
func Init(cfg Config) error {
	err := initSearch(cfg)
	if err != nil {
		return err
	}

	switch strings.ToLower(cfg.DatabaseType) {
	case "postgres":
		dbType = postgres
		err = initPostgres(cfg)
		if err != nil {
			return err
		}
	default:
		dbType = sqlite

		err = initSQLite(cfg)
		if err != nil {
			return err
		}
	}

	if cfg.MaxConnectionLifetime != "" {
		lifetime, err := time.ParseDuration(cfg.MaxConnectionLifetime)
		if err == nil {
			db.SetConnMaxLifetime(lifetime)
		} else {
			log.Printf("Invalid MaxConnectionLifetime duration format (%s), using default", cfg.MaxConnectionLifetime)
		}
	}
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetMaxOpenConns(cfg.MaxOpenConnections)

	err = ensureSchema(cfg.AllowSchemaRollback)
	if err != nil {
		return err
	}

	return nil
}

func initSQLite(cfg Config) error {
	url := ""
	if cfg.DatabaseFile == "" && cfg.DatabaseURL == "" {
		cfg.DatabaseFile = DefaultConfig().DatabaseFile
	}

	if cfg.DatabaseURL == "" {
		url = cfg.DatabaseFile
	}

	var err error
	db, err = sql.Open("sqlite3", url)
	if err != nil {
		return err
	}

	return nil
}

func initPostgres(cfg Config) error {
	// test if lexLibrary database exists and create if it doesn't
	return fmt.Errorf("TODO")
}
