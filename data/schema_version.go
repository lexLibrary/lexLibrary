// Copyright (c) 2017-2018 Townsourced Inc.

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
	|string    | nvarchar(size)    |
	|time.Time | timestamp/datetime|
	+------------------------------+


	Keep column and table names in lowercase and separate words with underscores
	tables should be named for their collections (i.e. plural)

	For best compatibility, only have one statement per version; i.e. no semicolons, and don't use any reserved words

	String / Text types will be by default case sensitive and unicode supported. The default database collations should
	reflect that.  Prefer Text over varchar except where necessary such as PKs.

	DateTime types are only precise up to Milliseconds

	Integers are 64 bit

	Add new versions for changes to exising tables if the changes have been checked into the Dev or master branches
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
				id {{varchar 20}} NOT NULL PRIMARY KEY,
				occurred {{datetime}} NOT NULL,
				message {{text}}
			)
		`),
		rollback: NewQuery("drop table logs"),
	},
	schemaVer{
		update:   NewQuery("create index i_occurred on logs (occurred)"),
		rollback: NewQuery("drop index i_occurred"),
	},
	schemaVer{
		update: NewQuery(`
			create table settings (
				id {{varchar 64}} NOT NULL PRIMARY KEY,
				description {{text}} NOT NULL,
				value {{text}} NOT NULL
			)
		`),
		rollback: NewQuery("DROP table settings"),
	},
	schemaVer{
		update: NewQuery(`
			create table users (
				id {{varchar 20}} NOT NULL PRIMARY KEY,
				username {{varchar 64}} NOT NULL,
				first_name {{text}},
				last_name {{text}},
				auth_type {{text}} NOT NULL,
				password {{bytes}},
				password_version {{int}},
				active {{bool}} NOT NULL,
				version {{int}} NOT NULL,
				updated {{datetime}} NOT NULL,
				created {{datetime}} NOT NULL
			)
		`),
		rollback: NewQuery("drop table users"),
	},
	schemaVer{
		update: NewQuery(`
			create table sessions (
				id {{varchar 32}} NOT NULL,
				user_id {{varchar 20}} NOT NULL,
				valid {{bool}},
				expires {{datetime}} NOT NULL,
				ip_address {{text}},
				user_agent {{text}},
				csrf_token {{text}} NOT NULL,
				updated {{datetime}} NOT NULL,
				created {{datetime}} NOT NULL,
				PRIMARY KEY(id, user_id)
			)
		`),
		rollback: NewQuery("drop table sessions"),
	},
	schemaVer{
		update:   NewQuery("create index i_username on users (username)"),
		rollback: NewQuery("drop index i_username"),
	},
	schemaVer{
		update:   NewQuery(`alter table users add admin {{bool}}`),
		rollback: NewQuery("alter table users drop column admin"),
	},
}
