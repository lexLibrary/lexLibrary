// Copyright (c) 2017 Townsourced Inc.

package data

import "fmt"

func ensureSchema() error {
	// NOTE: Not all DB's allow DDL in transactions, so this needs to run outside of one

	findSchemaTable := multiQuery{
		sqlite: "SELECT name FROM sqlite_master WHERE type='table' and name = 'schema_versions'",
		any:    "select table_name from information_schema.tables where table_name = 'schema_versions'",
	}

	// look up current schema version
	// if current schemaVer > database schemaVer
	// run the next schema version script
	// if current schemaVer < database schemaVer
	// return error unless force rollback

	rows, err := db.Query(findSchemaTable.String())
	if err != nil {
		return err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	ver := 0

	if rows.Next() {
		// table exists, lookup current version
	}

	for i := ver; ver < len(schemaVersions); i++ {
		db.Query(schemaVersions[i].update)
	}

	return fmt.Errorf("TODO")
}
