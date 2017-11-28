// Copyright (c) 2017 Townsourced Inc.

// Package data handles all the data and application structures handling for Lex Library
package data

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"path"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql" // register mysql
	_ "github.com/lib/pq"                        // register postgres
	_ "github.com/mattn/go-sqlite3"              // register sqlite3
	"github.com/pkg/errors"
)

// Database Types
const (
	sqlite      = iota // github.com/mattn/go-sqlite3
	postgres           // github.com/lib/pq
	mysql              // github.com/go-sql-driver/mysql/
	cockroachdb        // github.com/lib/pq
	tidb               // github.com/go-sql-driver/mysql/
)

const databaseName = "lex_library"

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
	case "mysql":
		dbType = mysql
		err = initMySQL(cfg)
	case "sqlite":
		dbType = sqlite
		err = initSQLite(cfg)
	default:
		return errors.New("Invalid database type")
	}
	if err != nil {
		return err
	}

	prepareQueries()

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

func testDB(attempt int) {
	maxAttempts := 10
	sleep := 3 * time.Second

	err := db.Ping()

	if err != nil {
		if attempt >= maxAttempts {
			log.Fatalf("Error Connecting to database: %s", err)
		}
		left := time.Duration(maxAttempts-attempt) * sleep
		log.Printf("Error Connecting to database: %s\n ... Retrying for %v. CTRL-c to stop", err, left)
		time.Sleep(sleep)
		testDB(attempt + 1)
	}
}

// Teardown cleanly tears down any data layer connections
func Teardown() error {
	log.Printf("Tearing down data connections")
	return db.Close()
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
	testDB(1)

	return nil
}

func initPostgres(cfg Config) error {
	var err error
	db, err = sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return err
	}
	testDB(1)

	dbName := ""

	err = db.QueryRow("SELECT current_database()").Scan(&dbName)
	if err != nil {
		return errors.Wrap(err, "Getting current database")
	}

	if dbName == "postgres" {
		// db connection is pointing at default database, check for lexLibrary DB
		// and create as necessary
		count := 0
		err = db.QueryRow("SELECT count(*) FROM pg_database WHERE datname = $1", databaseName).Scan(&count)
		if err != nil {
			return errors.Wrapf(err, "Looking for %s Database", databaseName)
		}

		if count == 0 {
			_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", databaseName))
			if err != nil {
				return errors.Wrapf(err, "Creating %s database", databaseName)
			}
		}
		// reopen DB connection on new database
		u, err := url.Parse(cfg.DatabaseURL)
		if err != nil {
			return err
		}

		u.Path = path.Join(u.Path, databaseName)
		db, err = sql.Open("postgres", u.String())
		if err != nil {
			return err
		}

		testDB(1)
	}
	// db connection is pointing at a specific database, use as lexLibrary DB

	return nil
}

func initMySQL(cfg Config) error {
	mCfg, err := mysqlDriver.ParseDSN(cfg.DatabaseURL)
	if err != nil {
		return err
	}

	mCfg.ParseTime = true

	db, err = sql.Open("mysql", mCfg.FormatDSN())
	if err != nil {
		return err
	}
	testDB(1)

	var dbName string

	err = db.QueryRow("SELECT IFNULL(DATABASE(),'mysql')").Scan(&dbName)
	if err != nil {
		return errors.Wrap(err, "Getting current database")
	}

	if dbName == "mysql" {
		// db connection is pointing at default database, check for lexLibrary DB
		// and create as necessary
		count := 0
		err = db.QueryRow(`
			SELECT count(*) 
			FROM INFORMATION_SCHEMA.SCHEMATA
			where SCHEMA_NAME = ?
		`, databaseName).Scan(&count)
		if err != nil {
			return errors.Wrapf(err, "Looking for %s Database", databaseName)
		}

		if count == 0 {
			_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", databaseName))
			if err != nil {
				return errors.Wrapf(err, "Creating %s database", databaseName)
			}
		}

		// reopen DB connection on new database

		mCfg.DBName = databaseName

		db, err = sql.Open("mysql", mCfg.FormatDSN())
		if err != nil {
			return err
		}

		testDB(1)
	}
	// db connection is pointing at a specific database, use as lexLibrary DB

	return nil
}
