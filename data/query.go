// Copyright (c) 2017 Townsourced Inc.

package data

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
)

// Query is a templated query that can run across
// multiple database backends
type Query struct {
	statement string
	args      []string
	tx        *sql.Tx
}

// NewQuery creates a new query from the template passed in
func NewQuery(tmpl string) *Query {
	q := &Query{}
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
			case mysql, tidb, sqlite:
				return "DATETIME"
			case postgres, cockroachdb:
				return "TIMESTAMP with time ZONE"
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

func (q *Query) orderedArgs(args []sql.NamedArg) []interface{} {
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

// Exec executes a templated query without returning any rows
func (q *Query) Exec(args ...sql.NamedArg) (sql.Result, error) {
	if q.tx != nil {
		return q.tx.Exec(q.statement, q.orderedArgs(args)...)
	}
	return db.Exec(q.statement, q.orderedArgs(args)...)
}

// Query executes a templated query that returns rows
func (q *Query) Query(args ...sql.NamedArg) (*sql.Rows, error) {
	if q.tx != nil {
		return q.tx.Query(q.statement, q.orderedArgs(args)...)
	}
	return db.Query(q.statement, q.orderedArgs(args)...)
}

// QueryRow executes a templated query that returns a single row
func (q *Query) QueryRow(args ...sql.NamedArg) *sql.Row {
	if q.tx != nil {
		return q.tx.QueryRow(q.statement, q.orderedArgs(args)...)
	}
	return db.QueryRow(q.statement, q.orderedArgs(args)...)
}

func (q *Query) copy() *Query {
	return &Query{
		statement: q.statement,
		args:      q.args,
		tx:        q.tx,
	}
}

// Tx returns a new copy of the query that runs in the passed in transaction
func (q *Query) Tx(tx *sql.Tx) *Query {
	copy := q.copy()
	copy.tx = tx
	return copy
}

// Statement returns the complied query template
func (q *Query) Statement() string {
	return q.statement
}

// BeginTx begins a transaction on the database
// If the function passed in returns an error, the transaction rolls back
// If it returns a nil error, then the transaction commits
func BeginTx(trnFunc func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = trnFunc(tx)
	if err != nil {
		rErr := tx.Rollback()
		if rErr != nil {
			return errors.Errorf("Error rolling back transaction.  Rollback error %s, Original error %s", rErr, err)
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "Error committing transaction")
	}

	return nil
}

// Debug runs the passed in query and returns a string of the results
// in a tab delimited format, with columns listed in the first row
// meant for debugging use. Will panic instead of throwing an error
func Debug(statement string, args ...sql.NamedArg) string {
	padding := 4
	result := ""
	q := NewQuery(statement)

	rows, err := q.Query(args...)
	if err != nil {
		panic(err)
	}

	columns, err := rows.Columns()
	if err != nil {
		panic(err)
	}
	types, err := rows.ColumnTypes()
	if err != nil {
		panic(err)
	}

	lengths := make([]int, len(columns))

	for i := range columns {
		lengths[i] = padding + len(columns[i])
		result += fmt.Sprintf("%"+strconv.Itoa(lengths[i])+"s", columns[i])
	}

	result += "\n"

	values := make([]interface{}, len(columns))
	for i := range columns {
		val := reflect.New(types[i].ScanType())
		values[i] = val.Interface()
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(values...)
		if err != nil {
			panic(err)
		}

		for i := range columns {
			result += fmt.Sprintf("%"+strconv.Itoa(lengths[i])+"v", values[i])[:lengths[i]]
		}
		result += "\n"
	}

	return result
}
