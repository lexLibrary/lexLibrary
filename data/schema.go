// Copyright (c) 2017 Townsourced Inc.

package data

import (
	"database/sql"
	"fmt"
	"log"
)

var schemaVersionInsert =queryTemplate("insert into schema_versions (version, rollback) values ({{?}}, {{?}})")

func ensureSchema(allowRollback bool) error {
	// NOTE: Not all DB's allow DDL in transactions, so this needs to run outside of one

	err := ensureSchemaTable()
	if err != nil {
		return err
	}

	fmt.Println("table created")

	return ensureSchemaVersion(allowRollback)
}

func ensureSchemaTable() error {
	findSchemaTable := multiQuery{
		sqlite: "SELECT name FROM sqlite_master WHERE type='table' and name = 'schema_versions'",
		other:  "select table_name from information_schema.tables where table_name = 'schema_versions'",
	}

	name := ""
	err := db.QueryRow(findSchemaTable.String()).Scan(&name)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		_, err = db.Query(schemaVersions[0].update)
		if err != nil {
			return err
		}

		_, err = db.Exec(schemaVersionInsert, 
		if err != nil {
			return err
		}
	}
	return nil
}

func ensureSchemaVersion(allowRollback bool) error {
	currentVer := len(schemaVersions) - 1

	dbVer := 0
	err := db.QueryRow("select version from schema_versions order by version desc limit 1").Scan(&dbVer)

	if err != nil {
		return err
	}

	if dbVer == currentVer {
		// server and database are on the same schema version
		return nil
	}

	if dbVer < currentVer {
		dbVer++
		log.Printf("Updating database schema to version %d", dbVer)
		_, err = db.Query(schemaVersions[dbVer].update)
		if err != nil {
			return err
		}

		_, err = db.Query(schemaVersionInsert,	dbVer, schemaVersions[dbVer].rollback)
		if err != nil {
			return err
		}

		return ensureSchemaVersion(allowRollback)
	}
	// check for forced rollback
	if allowRollback {
		log.Printf("Rolling back database schema to version %d", dbVer)
		rollback := ""
		err = db.QueryRow(queryTemplate("select rollback from schema_versions where version = {{?}}"), dbVer).
			Scan(&rollback)
		if err != nil {
			return err
		}

		_, err = db.Query(rollback)
		if err != nil {
			return err
		}

		_, err = db.Query(queryTemplate("delete from schema_versions where version = {{?}}"), dbVer)
		if err != nil {
			return err
		}
		return ensureSchemaVersion(allowRollback)
	}
	return fmt.Errorf("Database schema version (%d) is newer than the code schema version (%d)", dbVer, currentVer)

}
