// Copyright (c) 2017 Townsourced Inc.

package data

type schemaVer struct {
	update   *Query
	rollback *Query
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
	|string    | citext            |
	|string    | nvarchar(size)    |
	|time.Time | timestamp/datetime|
	+------------------------------+


	Keep column and table names in lowercase and separate words with underscores
	tables should be named for their collections (i.e. plural)

	For best compatibility, only have one statement per version; i.e. no semicolons, and don't use any reserved words

	String / Text types will be by default case sensitive and unicode supported. The default database collations should
	reflect that. Case Insensitive column types can be used in specific columns {{citext}}
*/

var schemaVersions = []schemaVer{
	schemaVer{
		update: NewQuery(`
			create table schema_versions (
				version INTEGER NOT NULL PRIMARY KEY,
				rollback_script {{text}} NOT NULL
			)
		`),
		rollback: NewQuery("drop table schema_versions"),
	},
	schemaVer{
		update: NewQuery(`
			create table logs (
				occurred {{datetime}} NOT NULL,
				message {{citext}}
			)
		`),
		rollback: NewQuery("DROP INDEX i_occurred ON logs"),
	},
	schemaVer{
		update:   NewQuery("create index i_occurred on logs (occurred)"),
		rollback: NewQuery("Drop index i_occurred"),
	},
	schemaVer{
		update: NewQuery(`
			create table settings (
				id {{varchar 256}} NOT NULL PRIMARY KEY, -- necessary for mysql PK
				description {{text}} NOT NULL,
				value {{text}} NOT NULL
			)
		`),
		rollback: NewQuery("DROP table settings"),
	},
}
