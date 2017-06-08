// Copyright (c) 2017 Townsourced Inc.
package data

import (
	"fmt"
	"time"
)

type Config struct {
	DatabaseFile string
	SearchFile   string

	PostGres *PostGresCFG
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

func Init(cfg Config) error {
	err := initSearch(cfg)
	if err != nil {
		return err
	}

	if cfg.PostGres != nil {
		err = initPostgres(cfg.PostGres)
		if err != nil {
			return err
		}
	} else {
		err = initSQLite(cfg.DatabaseFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func initSQLite(filename string) error {
	return fmt.Errorf("TODO")
}

func initPostgres(cfg *PostGresCFG) error {
	return fmt.Errorf("TODO")

}
