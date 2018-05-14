// Copyright (c) 2017-2018 Townsourced Inc.

package data

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
	"text/template"
	"time"

	"github.com/mattn/go-sqlite3" // register sqlite3
	"github.com/pkg/errors"
)

// Query is a templated query that can run across
// multiple database backends
type Query struct {
	template *template.Template
	args     []string
	tx       *sql.Tx
	data     interface{}
}

// Argument is a wrapper around sql.NamedArg so that a data behavior can be unified across all database backends
// mainly dateTime handling.  Always use the data.Arg function, and not the type directly
type Argument sql.NamedArg

// Arg defines an argument for use in a Lex Library query, and makes sure that data behaviors are consistent across
// multiple database backends
func Arg(name string, value interface{}) Argument {
	switch v := value.(type) {
	case time.Time:
		value = v.UTC()
	case NullTime:
		if v.Valid {
			v.Time = v.Time.UTC()
			value = v
		}
	}

	return Argument(sql.Named(name, value))
}

func (q *Query) orderedArgs(args []Argument) []interface{} {
	ordered := make([]interface{}, 0, len(q.args))

	for i := range q.args {
		for j := range args {
			if args[j].Name == q.args[i] {
				switch dbType {
				case postgres, cockroachdb, sqlserver:
					// named args
					ordered = append(ordered, sql.NamedArg(args[j]))
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

// NewQuery creates a new query from the template passed in
func NewQuery(tmpl string) *Query {
	q := &Query{}

	funcs := template.FuncMap{
		"arg": func(name string) string {
			// Args must be named and must be unique, and must use sql.Named
			if name == "" {
				panic("Arguments must be named in sql statements")
			}
			for i := range q.args {
				if name == q.args[i] {
					panic(fmt.Sprintf("%s already exists in the query arguments", name))
				}
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
		"bytes":    bytesColumn,
		"datetime": datetimeColumn,
		"text":     textColumn,
		"varchar":  varcharColumn,
		"id":       idColumn,
		"int":      intColumn,
		"bool":     boolColumn,
		"defaultDateTime": func() string {
			t := time.Time{}
			switch dbType {
			case mysql, mariadb:
				return t.Format("2006-01-02 15:04:05.000")
			case sqlite:
				return t.Format(sqlite3.SQLiteTimestampFormats[0])
			case postgres, cockroachdb:
				return t.Format(time.RFC3339)
			case sqlserver:
				return t.Format(time.RFC3339)
			default:
				panic("Unsupported database type")
			}
		},
		"NOW": func() string {
			t := time.Now().UTC()
			switch dbType {
			case mysql, mariadb:
				return fmt.Sprintf("'%s'", t.Format("2006-01-02 15:04:05.000"))
			case sqlite:
				return fmt.Sprintf("'%s'", t.Format(sqlite3.SQLiteTimestampFormats[0]))
			case postgres, cockroachdb:
				return fmt.Sprintf("'%s'", t.Format(time.RFC3339))
			case sqlserver:
				return fmt.Sprintf("'%s'", t.Format(time.RFC3339))
			default:
				panic("Unsupported database type")
			}
		},
		"TRUE": func() string {
			switch dbType {
			case mysql, mariadb, postgres, cockroachdb:
				return "true"
			case sqlite, sqlserver:
				return "1"
			default:
				panic("Unsupported database type")
			}
		},
		"FALSE": func() string {
			switch dbType {
			case mysql, mariadb, postgres, cockroachdb:
				return "false"
			case sqlite, sqlserver:
				return "0"
			default:
				panic("Unsupported database type")
			}
		},
		"db": DatabaseType,
		"sqlite": func() bool {
			return dbType == sqlite
		},
		"postgres": func() bool {
			return dbType == postgres
		},
		"mysql": func() bool {
			return dbType == mysql
		},
		"mariadb": func() bool {
			return dbType == mariadb
		},
		"cockroachdb": func() bool {
			return dbType == cockroachdb
		},
		"sqlserver": func() bool {
			return dbType == sqlserver
		},
		"limit": func() string {
			switch dbType {
			case sqlite, postgres, cockroachdb:
				return `"limit"`
			case mysql, mariadb:
				return "`limit`"
			case sqlserver:
				return "[limit]"
			default:
				panic("Unsupported database type")
			}
		},
	}

	t, err := template.New("").Funcs(funcs).Parse(tmpl)
	if err != nil {
		panic(fmt.Errorf("Error parsing query template '%s': %s", tmpl, err))
	}
	q.template = t
	return q
}

func (q *Query) execTemplate(args ...Argument) (string, error) {
	var b bytes.Buffer

	err := q.template.Execute(&b, struct {
		Args []Argument
		Data interface{}
	}{
		Args: args,
		Data: q.data,
	})

	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (q *Query) Data(data interface{}) *Query {
	nq := q.copy()
	nq.data = data
	return nq
}

// Exec executes a templated query without returning any rows
func (q *Query) Exec(args ...Argument) (sql.Result, error) {
	statement, err := q.execTemplate(args...)
	if err != nil {
		return nil, err
	}
	if q.tx != nil {
		return q.tx.Exec(statement, q.orderedArgs(args)...)
	}
	return db.Exec(statement, q.orderedArgs(args)...)
}

// Query executes a templated query that returns rows
func (q *Query) Query(args ...Argument) (*sql.Rows, error) {
	statement, err := q.execTemplate(args...)
	if err != nil {
		return nil, err
	}

	if q.tx != nil {
		return q.tx.Query(statement, q.orderedArgs(args)...)
	}
	return db.Query(statement, q.orderedArgs(args)...)
}

// Row is a wrapper around sql.Row so I can pass in an template execution error if one occurred
type Row struct {
	err error
	row *sql.Row
}

// Scan wraps the normal sql.Row.Scan function
func (r *Row) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	return r.row.Scan(dest...)
}

// QueryRow executes a templated query that returns a single row
func (q *Query) QueryRow(args ...Argument) *Row {
	statement, err := q.execTemplate(args...)
	if err != nil {
		return &Row{
			err: err,
		}
	}

	if q.tx != nil {
		return &Row{row: q.tx.QueryRow(statement, q.orderedArgs(args)...)}
	}
	return &Row{row: db.QueryRow(statement, q.orderedArgs(args)...)}
}

func (q *Query) copy() *Query {
	return &Query{
		template: q.template,
		args:     q.args,
		tx:       q.tx,
	}
}

// Tx returns a new copy of the query that runs in the passed in transaction if a transaction is passed in
// if tx is nil then the normal query is returned
func (q *Query) Tx(tx *sql.Tx) *Query {
	if tx == nil {
		return q
	}
	copy := q.copy()
	copy.tx = tx
	return copy
}

// Statement returns the complied query template
func (q *Query) Statement(args ...Argument) (string, error) {
	statement, err := q.execTemplate(args...)
	if err != nil {
		return "", err
	}

	return statement, nil
}

func (q *Query) String() string {
	statement, err := q.Statement()
	if err != nil {
		return fmt.Sprintf("Query could not be executed: %s", err)
	}
	return statement
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
func (q *Query) Debug(args ...Argument) string {
	padding := 25
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
func (q *Query) DebugPrint(args ...Argument) {
	fmt.Println(q.Debug(args...))
}
