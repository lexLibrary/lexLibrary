// Copyright (c) 2017 Townsourced Inc.

package data

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var queryBuildQueue []*Query

// Query is a templated query that can run across
// multiple database backends
type Query struct {
	statement string
	built     bool
	args      []string
	tx        *sql.Tx
}

// NewQuery creates a new query from the template passed in
func NewQuery(tmpl string) *Query {
	q := &Query{
		statement: tmpl,
	}

	if db != nil {
		q.buildTemplate()
	} else {
		queryBuildQueue = append(queryBuildQueue, q)
	}

	return q
}

func (q *Query) orderedArgs(args []sql.NamedArg) []interface{} {
	ordered := make([]interface{}, 0, len(q.args))

	for i := range q.args {
		for j := range args {
			if args[j].Name == q.args[i] {
				switch dbType {
				case postgres, cockroachdb, sqlserver:
					// named args
					ordered = append(ordered, args[j])
				default:
					// unnamed values
					ordered = append(ordered, args[j].Value)
				}
				break
			}
		}
	}
	return ordered
}

func (q *Query) buildTemplate() {
	if db == nil {
		panic("Can't build query templates before the database type is set")
	}
	funcs := template.FuncMap{
		"arg": func(name string) string {
			// Args must be named, and must use sql.Named
			if name == "" {
				panic("Arguments must be named in sql statements")
			}
			q.args = append(q.args, name)
			switch dbType {
			case postgres, cockroachdb:
				return "$" + strconv.Itoa(len(q.args))
			case sqlserver:
				return "@" + name
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
			case sqlserver:
				return "VARBINARY(max)"
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
			case sqlserver:
				return "DATETIME2"
			default:
				panic("Unsupported database type")
			}
		},
		"text": func() string {
			// case sensitive strings
			switch dbType {
			case sqlite, postgres, cockroachdb, mysql, tidb:
				return "TEXT"
			case sqlserver:
				return "nvarchar(max)"
			default:
				panic("Unsupported database type")
			}
		},
		"citext": func() string {
			// case in-sensitive strings
			switch dbType {
			case postgres:
				return "CITEXT"
			case cockroachdb:
				return "TEXT COLLATE en_u_ks_level2"
			case sqlite:
				return "TEXT COLLATE nocase"
			case mysql, tidb:
				return "TEXT CHARACTER SET utf8 COLLATE utf8_unicode_ci"
			case sqlserver:
				return "nvarchar(max) COLLATE Latin1_General_CI_AS"
			default:
				panic("Unsupported database type")
			}
		},
		"varchar": func(size int) string {
			switch dbType {
			case postgres, cockroachdb, mysql, tidb:
				return fmt.Sprintf("varchar(%d)", size)
			case sqlite:
				return "TEXT"
			case sqlserver:
				return fmt.Sprintf("nvarchar(%d)", size)
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
			case sqlserver:
				return "sqlserver"
			default:
				panic("Unsupported database type")
			}
		},
		"sqlite": func() bool {
			return dbType == sqlite
		},
		"postgres": func() bool {
			return dbType == postgres
		},
		"mysql": func() bool {
			return dbType == mysql
		},
		"cockroachdb": func() bool {
			return dbType == cockroachdb
		},
		"tidb": func() bool {
			return dbType == tidb
		},
		"sqlserver": func() bool {
			return dbType == sqlserver
		},
	}

	buff := bytes.NewBuffer([]byte{})
	err := template.Must(template.New("").Funcs(funcs).Parse(q.statement)).Execute(buff, nil)
	if err != nil {
		panic(fmt.Errorf("Error building query template: %s", err))
	}

	q.statement = strings.TrimSpace(buff.String())
	q.built = true
}

// Exec executes a templated query without returning any rows
func (q *Query) Exec(args ...sql.NamedArg) (sql.Result, error) {
	if !q.built {
		q.buildTemplate()
	}
	if q.tx != nil {
		return q.tx.Exec(q.statement, q.orderedArgs(args)...)
	}
	return db.Exec(q.statement, q.orderedArgs(args)...)
}

// Query executes a templated query that returns rows
func (q *Query) Query(args ...sql.NamedArg) (*sql.Rows, error) {
	if !q.built {
		q.buildTemplate()
	}
	if q.tx != nil {
		return q.tx.Query(q.statement, q.orderedArgs(args)...)
	}
	return db.Query(q.statement, q.orderedArgs(args)...)
}

// QueryRow executes a templated query that returns a single row
func (q *Query) QueryRow(args ...sql.NamedArg) *sql.Row {
	if !q.built {
		q.buildTemplate()
	}
	if q.tx != nil {
		return q.tx.QueryRow(q.statement, q.orderedArgs(args)...)
	}
	return db.QueryRow(q.statement, q.orderedArgs(args)...)
}

func (q *Query) copy() *Query {
	return &Query{
		statement: q.statement,
		args:      q.args,
		built:     q.built,
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
	if !q.built {
		q.buildTemplate()
	}
	return q.statement
}

func (q *Query) String() string {
	return q.Statement()
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
func (q *Query) Debug(args ...sql.NamedArg) string {
	padding := 20
	result := ""

	rows, err := q.Query(args...)
	if err != nil {
		panic(err)
	}

	columns, err := rows.Columns()
	if err != nil {
		panic(err)
	}

	lengths := make([]int, len(columns))

	wrap := ""
	cols := ""
	for i := range columns {
		lengths[i] = padding + len(columns[i])
		cols += fmt.Sprintf("%-"+strconv.Itoa(lengths[i])+"s", columns[i])
		for j := 0; j < lengths[i]; j++ {
			wrap += "-"
		}
	}

	result += wrap + "\n" + cols + "\n" + wrap + "\n"

	values := make([]interface{}, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	count := 0
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err)
		}

		for i := range columns {
			var str string
			switch values[i].(type) {
			case nil:
				str = "NULL"
			case []byte:
				str = string(values[i].([]byte))
			default:
				str = fmt.Sprintf("%v", values[i])
			}

			val := fmt.Sprintf("%-"+strconv.Itoa(lengths[i])+"s", str)
			if i != len(columns)-1 {
				// don't trim last column
				val = val[:lengths[i]-3] + "..."
			}
			result += val

		}
		result += "\n"
		count++
	}

	return result + wrap + "\n(" + strconv.Itoa(count) + " rows)\n"
}

// DebugPrint prints out the debug query to the screen
func (q *Query) DebugPrint(args ...sql.NamedArg) {
	fmt.Println(q.Debug(args...))
}

func prepareQueries() error {
	for i := range queryBuildQueue {
		queryBuildQueue[i].buildTemplate()
	}
	return nil
}
