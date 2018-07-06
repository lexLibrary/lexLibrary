// Copyright (c) 2017-2018 Townsourced Inc.

package data

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

var schemaVersions = []*Query{
	NewQuery(`
		create table schema_versions (
			version {{int}} PRIMARY KEY NOT NULL,
			script {{text}} NOT NULL,
			occurred {{datetime}} NOT NULL
		)
	`),
	NewQuery(`
		create table logs (
			id {{id}} PRIMARY KEY NOT NULL,
			occurred {{datetime}} NOT NULL, 
			message {{text}} NOT NULL
		)
	`),
	NewQuery(`
		create index i_occurred on logs (occurred)
	`),
	NewQuery(`
		create table settings (
			id {{varchar 64}} PRIMARY KEY NOT NULL,
			description {{text}} NOT NULL,
			value {{text}}  NOT NULL
		)
	`),
	NewQuery(`
		create table users (
			id {{id}} PRIMARY KEY NOT NULL,
			username {{varchar "user.username"}} NOT NULL,
			name {{varchar "user.name"}},
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
	`),
	NewQuery(`
		create table sessions (
			id {{varchar 32}} NOT NULL,
			user_id {{id}} NOT NULL REFERENCES users(id),
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
	`),
	NewQuery(`
		create index i_username on users (username)
	`),
	NewQuery(`
		create table images (
			id {{id}} PRIMARY KEY NOT NULL,
			name {{text}} NOT NULL,
			version {{int}} NOT NULL,
			content_type {{text}} NOT NULL,
			data {{bytes}} NOT NULL,
			thumb {{bytes}} NOT NULL,
			placeholder {{bytes}} NOT NULL,
			updated {{datetime}} NOT NULL,
			created {{datetime}} NOT NULL
		)	
	`),
	NewQuery(`
		{{if cockroachdb}}
			alter table users add column profile_image_id {{id}};
			CREATE INDEX ON users (profile_image_id);
			alter table users add foreign key (profile_image_id) references images(id);
		{{else}}
			alter table users add profile_image_id {{id}} REFERENCES images(id)
		{{end}}
	`),
	NewQuery(`
		{{if cockroachdb}}
			alter table users add column profile_image_draft_id {{id}};
			CREATE INDEX ON users (profile_image_draft_id);
			alter table users add foreign key (profile_image_draft_id) references images(id);
		{{else}}
			alter table users add profile_image_draft_id {{id}} REFERENCES images(id)
		{{end}}
	`),
	NewQuery(`
		create table groups (
			id {{id}} PRIMARY KEY NOT NULL,
			name {{varchar "group.name"}} UNIQUE NOT NULL,
			version {{int}} NOT NULL,
			updated {{datetime}} NOT NULL,
			created {{datetime}} NOT NULL
		)	
	`),
	NewQuery(`
		create table group_users (
			user_id {{id}} NOT NULL REFERENCES users(id),
			group_id {{id}} NOT NULL REFERENCES groups(id),
			admin {{bool}},
			PRIMARY KEY(user_id, group_id)
		)
	`),
	NewQuery(`
		create index i_name on groups (name)
	`),
	NewQuery(`
		create table registration_tokens (
			token {{varchar 32}} PRIMARY KEY NOT NULL,
			{{limit}} {{int}} NOT NULL,
			expires {{datetime}},
			valid {{bool}} NOT NULL,
			updated {{datetime}} NOT NULL,
			created {{datetime}} NOT NULL,
			creator {{id}} NOT NULL
		)
	`),
	NewQuery(`
		create table registration_token_groups (
			token {{varchar 32}} NOT NULL REFERENCES registration_tokens(token),
			group_id {{id}} NOT NULL REFERENCES groups(id),
			PRIMARY KEY(token, group_id)
		)
	`),
	NewQuery(`
		create table registration_token_users (
			token {{varchar 32}} NOT NULL REFERENCES registration_tokens(token),
			user_id {{id}} NOT NULL REFERENCES users(id),
			PRIMARY KEY(token, user_id)
		)
	`),
	NewQuery(`
		alter table registration_tokens add description {{text}}
	`),
	NewQuery(`
		alter table groups add name_search {{varchar "group.name"}}
	`),
	NewQuery(`
		create index i_name_search on groups (name_search)
	`),
	NewQuery(`
		create table documents (
			id {{id}} PRIMARY KEY NOT NULL,
			title {{text}} NOT NULL,
			content {{text}} NOT NULL,
			version {{int}} NOT NULL,
			draft_id {{id}} NOT NULL,
			updated {{datetime}} NOT NULL,
			created {{datetime}} NOT NULL,
			creator {{id}} NOT NULL REFERENCES users(id), 
			updater {{id}} NOT NULL REFERENCES users(id)
		)
	`),
	NewQuery(`
		create table document_groups (
			document_id {{id}} NOT NULL REFERENCES documents(id),
			group_id {{id}} NOT NULL REFERENCES groups(id),
			PRIMARY KEY(document_id, group_id)
		)
	`),
	NewQuery(`
		create table document_drafts (
			id {{id}} PRIMARY KEY NOT NULL,
			document_id {{id}} REFERENCES documents(id),
			title {{text}} NOT NULL,
			content {{text}} NOT NULL,
			version {{int}} NOT NULL,
			updated {{datetime}} NOT NULL,
			created {{datetime}} NOT NULL,
			creator {{id}} NOT NULL REFERENCES users(id), 
			updater {{id}} NOT NULL REFERENCES users(id)
		)
	`),
	NewQuery(`
		create table document_history (
			document_id {{id}} NOT NULL REFERENCES documents(id),
			version {{int}} NOT NULL,
			draft_id {{id}} NOT NULL,
			title {{text}} NOT NULL,
			content {{text}} NOT NULL,
			created {{datetime}} NOT NULL,
			creator {{id}} NOT NULL REFERENCES users(id),
			PRIMARY KEY(document_id, version)
		)
	`),
	NewQuery(`
		create table document_tags (
			document_id {{id}} NOT NULL REFERENCES documents(id),
			tag {{varchar "document.tag"}} NOT NULL, 
			type {{text}} NOT NULL,
			stem {{text}},
			PRIMARY KEY(document_id, tag)
		)
	`),
	NewQuery(`
		create table document_draft_tags (
			draft_id {{id}} NOT NULL REFERENCES document_drafts(id),
			tag {{varchar "document.tag"}} NOT NULL, 
			type {{text}} NOT NULL,
			stem {{text}},
			PRIMARY KEY(draft_id, tag)
		)
	`),
}
