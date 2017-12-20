// Copyright (c) 2017 Townsourced Inc.

package data

import (
	"database/sql"
	"log"

	"github.com/pkg/errors"
)

var schemaVersionInsert = NewQuery(`insert into schema_versions (version, rollback_script) 
	values ({{arg "version"}}, {{arg "rollback"}})`)

func ensureSchema(allowRollback bool) error {
	// NOTE: Not all DB's allow DDL in transactions, so this needs to run outside of one

	err := ensureSchemaTable()
	if err != nil {
		return err
	}

	return ensureSchemaVersion(allowRollback)
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
		_, err = schemaVersions[0].update.Exec()
		if err != nil {
			return errors.Wrap(err, "Creating schema_versions table")
		}

		_, err := schemaVersionInsert.Exec(
			sql.Named("version", 0),
			sql.Named("rollback", schemaVersions[0].rollback.Statement()))
		if err != nil {
			return errors.Wrap(err, "Inserting first schema version")
		}
	}
	return nil
}

func ensureSchemaVersion(allowRollback bool) error {
	currentVer := len(schemaVersions) - 1

	dbVer := 0
	err := db.QueryRow(`select max(version) from schema_versions`).Scan(&dbVer)
	if err == sql.ErrNoRows {
		_, err := schemaVersionInsert.Exec(
			sql.Named("version", 0),
			sql.Named("rollback", schemaVersions[0].rollback.Statement()))
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
		_, err = schemaVersions[dbVer].update.Exec()
		if err != nil {
			return errors.Wrapf(err, "Updating schema to version %d", dbVer)
		}

		_, err = schemaVersionInsert.Exec(
			sql.Named("version", dbVer),
			sql.Named("rollback", schemaVersions[dbVer].rollback.Statement()))
		if err != nil {
			return errors.Wrapf(err, "Inserting schema version %d", dbVer)
		}

		return ensureSchemaVersion(allowRollback)
	}
	// check for forced rollback
	if allowRollback {
		log.Printf("Rolling back database schema to version %d", dbVer)
		rollback := ""
		err = NewQuery(`select rollback_script from schema_versions where version = {{arg "version"}}`).QueryRow(
			sql.Named("version", dbVer)).Scan(&rollback)
		if err != nil {
			return errors.Wrapf(err, "Looking for rollback script for version %d", dbVer)
		}

		_, err = db.Exec(rollback)
		if err != nil {
			return errors.Wrapf(err, "Executing rollback script for version %d", dbVer)
		}

		_, err = NewQuery(`delete from schema_versions where version = {{arg "version"}}`).Exec(
			sql.Named("version", dbVer))
		if err != nil {
			return errors.Wrapf(err, "Removing schema version from database for version %d", dbVer)
		}
		return ensureSchemaVersion(allowRollback)
	}
	return errors.Errorf("Database schema version (%d) is newer than the code schema version (%d)", dbVer, currentVer)

}
