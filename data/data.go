// Copyright (c) 2017 Townsourced Inc.
package data

import "time"

type Config struct {
	SQLiteFile string
	SearchFile string

	PostGres PostGresCFG
}

type PostGresCFG struct {
	DBname         string
	User           string
	Password       string
	Host           string
	Port           string
	SSLMode        string
	ConnectTimeout time.Duration
	SSLCert        string
	SSLKey         string
	SSLRootCert    string
}

func Init(cfg Config) {

}

func initSQLite(filename string) {

}

func initSearch(filename string) {

}

func initPostgres(cfg PostGresCFG) {

}
