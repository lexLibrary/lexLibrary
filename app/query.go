// Copyright (c) 2017 Townsourced Inc.

package app

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
)

type query struct {
	statement string
	args      []string
}

func newQuery(tmpl string) *query {
	q := &query{}
	funcs := template.FuncMap{
		"arg": func(name string) string {
			// Args must be named, and must use sql.Named
			if name == "" {
				panic("Arguments must be named in sql statements")
			}
			q.args = append(q.args, name)
			switch dbType {
			case postgres, cockroachdb:
				return "$" + name
			default:
				return "?"
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

	buff := bytes.NewBuffer([]byte{})
	err := template.Must(template.New("").Funcs(funcs).Parse(tmpl)).Execute(buff, nil)
	if err != nil {
		panic(fmt.Errorf("Error building query template: %s", err))
	}

	q.statement = buff.String()
	return q
}

func (q *query) orderedArgs(args []sql.NamedArg) []interface{} {
	ordered := make([]interface{}, 0, len(q.args))

	for i := range q.args {
		for j := range args {
			if args[j].Name == q.args[i] {
				switch dbType {
				case postgres, cockroachdb:
					ordered = append(ordered, args[j])
				default:
					ordered = append(ordered, args[j].Value)
				}
				break
			}
		}
	}
	return ordered
}

func (q *query) exec(args ...sql.NamedArg) (sql.Result, error) {
	return db.Exec(q.statement, q.orderedArgs(args)...)
}

func (q *query) query(args ...sql.NamedArg) (*sql.Rows, error) {
	return db.Query(q.statement, q.orderedArgs(args)...)
}
func (q *query) queryRow(args ...sql.NamedArg) *sql.Row {
	return db.QueryRow(q.statement, q.orderedArgs(args)...)
}

//TODO: Handle transactions
// TODO: Handle timeouts
