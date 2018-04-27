// Copyright (c) 2017-2018 Townsourced Inc.

package data

import (
	"database/sql"
	"log"

	"github.com/pkg/errors"
)

var schemaVersionInsert = NewQuery(`
	insert into schema_versions (version, script, occurred) values ({{arg "version"}}, {{arg "script"}}, {{NOW}})
`)

func ensureSchema() error {
	// NOTE: Not all DB's allow DDL in transactions, so this needs to run outside of one

	err := ensureSchemaTable()
	if err != nil {
		return errors.Wrap(err, "Ensuring schema table")
	}

	err = ensureSchemaVersion()
	if err != nil {
		return errors.Wrap(err, "Ensuring schema version")
	}
	return nil
}

func ensureSchemaTable() error {
	findSchemaTable := NewQuery(`
		{{if sqlite}}
			SELECT name FROM sqlite_master WHERE type = 'table' and name = 'schema_versions'
		{{else}}
			select table_name from INFORMATION_SCHEMA.TABLES where table_name = 'schema_versions'
		{{end}}
	`)

	name := ""
	err := findSchemaTable.QueryRow().Scan(&name)
	if err != nil {
		if err != sql.ErrNoRows {
			return errors.Wrap(err, "Looking for schema_versions table")
		}
		_, err = schemaVersions[0].Exec()
		if err != nil {
			return errors.Wrap(err, "Creating schema_versions table")
		}

		_, err := schemaVersionInsert.Exec(Arg("version", 0), Arg("script", schemaVersions[0].Statement()))
		if err != nil {
			return errors.Wrap(err, "Inserting first schema version")
		}
	}
	return nil
}

func ensureSchemaVersion() error {
	currentVer := len(schemaVersions) - 1

	dbVer := 0
	err := db.QueryRow(`select max(version) from schema_versions`).Scan(&dbVer)
	if err == sql.ErrNoRows {
		_, err := schemaVersionInsert.Exec(Arg("version", 0), Arg("script", schemaVersions[0].Statement()))
		if err != nil {
			return errors.Wrap(err, "Inserting first schema version")
		}
	} else if err != nil {
		return errors.Wrap(err, "Getting current schema version from database")
	}

	if dbVer == currentVer {
		// server and database are on the same schema version
		return nil
	}

	if dbVer < currentVer {
		dbVer++
		log.Printf("Updating database schema to version %d", dbVer)
		_, err = schemaVersions[dbVer].Exec()
		if err != nil {
			return errors.Wrapf(err, "Updating schema to version %d", dbVer)
		}

		_, err = schemaVersionInsert.Exec(Arg("version", dbVer), Arg("script", schemaVersions[dbVer].Statement()))
		if err != nil {
			return errors.Wrapf(err, "Inserting schema version %d", dbVer)
		}

		return ensureSchemaVersion()
	}

	return errors.Errorf("Database schema version (%d) is newer than the code schema version (%d)", dbVer, currentVer)
}
