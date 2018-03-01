// Copyright (c) 2017-2018 Townsourced Inc.

package data

import "github.com/pkg/errors"

type schemaVer struct {
	updates []*Query
}

func (s schemaVer) exec() error {
	for i := range s.updates {
		_, err := s.updates[i].Exec()
		if err != nil {
			return errors.Wrapf(err, "Executing update #%d", i)
		}
	}
	return nil
}

func newSchemaVer(templates ...string) schemaVer {
	updates := make([]*Query, len(templates))

	for i := range templates {
		updates[i] = NewQuery(templates[i])
	}

	return schemaVer{
		updates: updates,
	}
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

	For best compatibility, only have one statement per query; i.e. no semicolons, and don't use any reserved words

	String / Text types will be by default case sensitive and unicode supported. The default database collations should
	reflect that.  Prefer Text over varchar except where necessary such as PKs.

	DateTime types are only precise up to Milliseconds

	Integers are 64 bit

	Add new versions for changes to exising tables if the changes have been checked into the Dev or master branches

	When in doubt, define the database tables to match the behavior of Go.  For instance, no null / nil types on
	strings, datetimes, booleans, etc.  Strings should be case sensitive, because they are in Go.  An unset time
	is it's Zero value which should be the column default.
*/

var schemaVersions = []schemaVer{
	newSchemaVer(`
		create table schema_versions (
			version INTEGER PRIMARY KEY NOT NULL
		)
	`),
	newSchemaVer(`
		create table logs (
			id {{varchar 20}} PRIMARY KEY NOT NULL,
			occurred {{datetime}} NOT NULL, 
			message {{text}} NOT NULL
		)
	`, `
		create index i_occurred on logs (occurred)
	`, `
		create table settings (
			id {{varchar 64}} PRIMARY KEY NOT NULL,
			description {{text}} NOT NULL,
			value {{text}}  NOT NULL
		)
	`),
	newSchemaVer(`
		create table users (
			id {{varchar 20}} PRIMARY KEY NOT NULL,
			username {{varchar 64}} NOT NULL,
			first_name {{text}},
			last_name {{text}},
			auth_type {{text}} NOT NULL,
			password {{bytes}},
			password_version {{int}},
			password_expiration {{datetime}},
			active {{bool}},
			admin {{bool}} NOT NULL,
			version {{int}} NOT NULL,
			updated {{datetime}} NOT NULL,
			created {{datetime}} NOT NULL
		)
	`, `
		create table sessions (
			id {{varchar 32}} NOT NULL,
			user_id {{varchar 20}} NOT NULL REFERENCES users(id),
			valid {{bool}} NOT NULL,
			expires {{datetime}} NOT NULL,
			ip_address {{text}} NOT NULL,
			user_agent {{text}},
			csrf_token {{text}} NOT NULL,
			csrf_date {{datetime}} NOT NULL,
			updated {{datetime}} NOT NULL,
			created {{datetime}} NOT NULL,
			PRIMARY KEY(id, user_id)
		)
	`, `
		create index i_username on users (username)
	`),
}
