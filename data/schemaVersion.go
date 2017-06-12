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
*/

var schemaVersions = []schemaVer{
	schemaVer{
		update: `
		create table schema_versions (
			version INTEGER NOT NULL PRIMARY KEY,
			rollback text NOT NULL
		);
		insert into schema_versions (version, rollback) values (0, 'drop table schema_versions');
		`,
		rollback: "drop table schema_versions",
	},
}
