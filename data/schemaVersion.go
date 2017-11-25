// Copyright (c) 2017 Townsourced Inc.

package data

type schemaVer struct {
	update   string
	rollback string
}

/*
	Array index determines the schema version
	Add new schema versions to the bottom of the array
	Once you push your changes to the main git repository, add new schema versions for
	table changes, rather than updating existing ones

	Stick to the following column types:

	+------------------------------+
	|go        | sql type          |
	|----------|-------------------|
	|nil       | null              |
	|int       | integer           |
	|int64     | integer           |
	|float64   | float             |
	|bool      | integer           |
	|[]byte    | blob              |
	|string    | text              |
	|time.Time | timestamp/datetime|
	+------------------------------+


	Keep column and table names in lowercase and separate words with underscores
	tables should be named for their collections (i.e. plural)
*/

var schemaVersions = []schemaVer{
	schemaVer{
		update: NewQuery(`
			create table schema_versions (
				version INTEGER NOT NULL PRIMARY KEY,
				rollback {{text}} NOT NULL
			);
		`).Statement(),
		rollback: "drop table schema_versions",
	},
	schemaVer{
		update: NewQuery(`
			create table logs (
				occurred {{datetime}} NOT NULL,
				message {{text}}
			);
			create index i_occurred on logs (occurred);
		`).statement,
		rollback: "drop table logs",
	},
}
