// Copyright (c) 2017-2018 Townsourced Inc.

// Package data handles all the data connections and management for Lex Library
package data

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"net/url"
	"path"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"         // register sqlserver
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
	sqlserver          // github.com/denisenkom/go-mssqldb
	cockroachdb        // github.com/lib/pq
)

const databaseName = "lex_library"

var (
	db         *sql.DB
	dbType     int
	currentCFG Config
)

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
}

// DefaultConfig returns the default configuration for the data layer
func DefaultConfig() Config {
	return Config{
		DatabaseType: "sqlite",
		DatabaseFile: "./lexLibrary.db",
		SearchFile:   "./lexLibrary.search",
	}
}

// Init initializes the data layer based on the passed in configuration.
// Initialization includes things like setting up the database and the connections to it.
func Init(cfg Config) error {
	log.Printf("Initializing Data Layer on %s", cfg.DatabaseType)
	err := initSearch(cfg)
	if err != nil {
		return err
	}

	switch strings.ToLower(cfg.DatabaseType) {
	case "postgres":
		dbType = postgres
		err = initPostgresAndCDB(cfg)
	case "mysql", "mariadb":
		dbType = mysql
		err = initMySQL(cfg)
	case "sqlite":
		dbType = sqlite
		err = initSQLite(cfg)
	case "cockroachdb":
		dbType = cockroachdb
		err = initPostgresAndCDB(cfg)
	case "sqlserver":
		dbType = sqlserver
		err = initSQLServer(cfg)
	default:
		return errors.New("Invalid database type")
	}
	if err != nil {
		return err
	}

	err = prepareQueries()
	if err != nil {
		return err
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

	currentCFG = cfg

	return ensureSchema()
}

// CurrentCFG returns the current config LL is running on
func CurrentCFG() Config {
	return currentCFG
}

func testDB(attempt int) {
	maxAttempts := 100
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
	if db != nil {
		return db.Close()
	}
	return nil
}

func initSQLite(cfg Config) error {
	url := cfg.DatabaseURL

	if cfg.DatabaseFile == "" && cfg.DatabaseURL == "" {
		cfg.DatabaseFile = DefaultConfig().DatabaseFile
	}

	if url == "" {
		url = cfg.DatabaseFile
	}

	if cfg.MaxOpenConnections == 1 {
		return errors.New("MaxOpenConnections must be more than 1 for sqlite")
	}

	var err error
	db, err = sql.Open("sqlite3", url)
	if err != nil {
		return err
	}
	testDB(1)

	return nil
}

func initPostgresAndCDB(cfg Config) error {
	var err error
	db, err = sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return err
	}

	testDB(1)

	dbName := ""

	err = db.QueryRow("SELECT COALESCE(current_database(), 'postgres')").Scan(&dbName)
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

	return err
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
			_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s character set utf8 collate utf8_bin", databaseName))
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

func initSQLServer(cfg Config) error {
	var err error
	db, err = sql.Open("sqlserver", cfg.DatabaseURL)
	if err != nil {
		return err
	}

	testDB(1)

	dbName := ""

	err = db.QueryRow("SELECT DB_NAME()").Scan(&dbName)
	if err != nil {
		return errors.Wrap(err, "Getting current database")
	}

	if dbName == "master" {
		// db connection is pointing at default database, check for lexLibrary DB
		// and create as necessary
		count := 0
		err = db.QueryRow("SELECT count(*) FROM master.dbo.sysdatabases WHERE [name] = @name",
			sql.Named("name", databaseName)).Scan(&count)
		if err != nil {
			return errors.Wrapf(err, "Looking for %s Database", databaseName)
		}

		if count == 0 {
			_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s COLLATE Latin1_General_CS_AS", databaseName))
			if err != nil {
				return errors.Wrapf(err, "Creating %s database", databaseName)
			}
		}
		// reopen DB connection on new database
		u, err := url.Parse(cfg.DatabaseURL)
		if err != nil {
			return err
		}

		val := url.Values{}
		val.Set("database", databaseName)
		u.RawQuery = val.Encode()

		db, err = sql.Open("sqlserver", u.String())
		if err != nil {
			return err
		}

		testDB(1)
	}
	// db connection is pointing at a specific database, use as lexLibrary DB

	return nil
}

// NullTime represents a time.Time that may be null. NullTime implements the
// sql.Scanner interface so it can be used as a scan destination, similar to
// sql.NullString.
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Scan implements the Scanner interface.
func (nt *NullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	return nil
}

// Value implements the driver Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// MarshalJSON implements the JSON interface for NullTime
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return nt.Time.MarshalJSON()
}

// UnmarshalJSON implements the JSON interface for NullTime
func (nt *NullTime) UnmarshalJSON(data []byte) error {
	if data == nil {
		nt.Valid = false
	}
	return nt.Time.UnmarshalJSON(data)
}
