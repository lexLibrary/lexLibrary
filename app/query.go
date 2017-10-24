// Copyright (c) 2017 Townsourced Inc.

package app

import (
	"bytes"
	"html/template"
	"strconv"
	"database/sql"
)

type queryTemplate {
	template.Template
	args []sql.NamedArg
}

/* query is a database query template that contains multiple definitions for different database backend
for use with databases where a query can't be shared across all DB backends
*/
func query(tmpl string) *queryTemplate {
	buff := bytes.NewBuffer([]byte{})

	funcs := template.FuncMap{
		"arg": func(name string) string {
			switch dbType {
			case postgres, cockroachdb:
				return "$" + name
			default:
				return "?"
			}
		},
		"offsetLimit": func(offset, limit int) string {
			// FIXME: offset limit is a different order for different database backends
			switch dbType {
			case sqlite:
				return "LIMIT " + strconv.Itoa(limit) + " OFFSET " + strconv.Itoa(offset)
			default:
				return "OFFSET " + strconv.Itoa(offset) + " ROWS FETCH NEXT " + strconv.Itoa(limit) +
					" ROWS ONLY"
			}
		},
		"bytes": func() string {
			switch dbType {
			case sqlite:
				return "BLOB"
			case postgres:
				return "BYTEA"
			case cockroachdb:
				return "BYTES"
			case mysql, tidb:
				return "VARBINARY"
			default:
				panic("Unsupported database type")
			}
		},
		"datetime": func() string {
			switch dbType {
			case sqlite:
				return "TEXT"
			case postgres, cockroachdb:
				return "TIMESTAMP with time ZONE"
			case mysql, tidb:
				return "DATETIME"
			default:
				panic("Unsupported database type")
			}
		},
		"text": func() string {
			switch dbType {
			case sqlite, postgres, cockroachdb, mysql, tidb:
				return "TEXT"
			default:
				panic("Unsupported database type")
			}
		},
		"db": func() string {
			switch dbType {
			case sqlite:
				return "sqlite"
			case postgres:
				return "postgres"
			case mysql:
				return "mysql"
			case cockroachdb:
				return "cockroachdb"
			case tidb:
				return "tidb"
			default:
				panic("Unsupported database type")
			}
		},
		"sqlite": func() bool {
			if dbType == sqlite {
				return true
			}
			return false
		},
		"postgres": func() bool {
			if dbType == postgres {
				return true
			}
			return false
		},
		"mysql": func() bool {
			if dbType == mysql {
				return true
			}
			return false
		},
		"cockroachdb": func() bool {
			if dbType == cockroachdb {
				return true
			}
			return false
		},
		"tidb": func() bool {
			if dbType == tidb {
				return true
			}
			return false
		},
	}

	t := template.Must(template.New("").Funcs(funcs).Parse(tmpl))

	return (*queryTemplate)(t)
}
