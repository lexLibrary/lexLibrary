// Copyright (c) 2017 Townsourced Inc.

package app

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
)

/* query is a query that contains multiple definitions for different database backend
for use with databases where a query can't be shared across all DB backends

	q := multiQuery{
		sqlite:   "select * from tbl LIMIT 10 OFFSET 50",
		other: "select * from tbl OFFSET 50 ROWS FETCH NEXT 10 ROWS ONLY",
	}

	fmt.Println(q.String())
*/
type query struct {
	sqlite      string
	postgres    string
	mysql       string
	cockroachdb string
	tidb        string
	other       string // other will be used where a query isn't specified

	limit  int
	offset int
}

func (q query) String() string {
	switch dbType {
	case sqlite:
		if q.sqlite != "" {
			return q.sqlite
		}
		return q.other
	case postgres:
		if q.postgres != "" {
			return q.postgres
		}
		return q.other
	case cockroachdb:
		if q.cockroachdb != "" {
			return q.cockroachdb
		}
		return q.other

	case mysql:
		if q.mysql != "" {
			return q.mysql
		}
		return q.other
	case tidb:
		if q.tidb != "" {
			return q.tidb
		}
		return q.other
	default:
		panic("Unsupported database type")
	}
}

// queryTemplate is for easily creating queries that can run against multiple database backends
func queryTemplate(tmpl string) string {
	buff := bytes.NewBuffer([]byte{})
	paramCount := 0

	funcs := template.FuncMap{
		"param": func() string {
			paramCount++
			switch dbType {
			case postgres, cockroachdb:
				return "$" + strconv.Itoa(paramCount)
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
	}

	err := template.Must(template.New("").Funcs(funcs).Parse(tmpl)).Execute(buff, nil)
	if err != nil {
		panic(fmt.Sprintf("Error build query template: %s", err))
	}

	return buff.String()
}
