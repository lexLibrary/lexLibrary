// Copyright (c) 2017 Townsourced Inc.

package app

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
*/

var schemaVersions = []schemaVer{
	schemaVer{
		update: queryTemplate(`
		create table schema_versions (
			version INTEGER NOT NULL PRIMARY KEY,
			rollback {{text}} NOT NULL
		);
		`),
		rollback: "drop table schema_versions",
	},
	schemaVer{
		update: queryTemplate(`
		create table logs (
			occurred {{datetime}} NOT NULL,
			message {{text}}
		);
		create index i_occurred on logs (occurred);
		`),
		rollback: `drop table logs`,
	},
}
