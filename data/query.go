// Copyright (c) 2017 Townsourced Inc.

package data

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
)

// multiQuery is query that contains multiple definitions for different database backend
// for use with databases where a query can't be shared across all DB backends
/*
	q := multiQuery{
		sqlite:   "select * from tbl LIMIT 10 OFFSET 50",
		postgres: "select * from tbl OFFSET 50 ROWS FETCH NEXT 10 ROWS ONLY",
	}

	fmt.Println(q.String())
*/
type multiQuery struct {
	sqlite      string
	postgres    string
	mysql       string
	cockroachdb string
	tidb        string
	any         string // any will be used where a query isn't specified
}

func (q multiQuery) String() string {
	switch dbType {
	case sqlite:
		if q.sqlite != "" {
			return q.sqlite
		}
		return q.any
	case postgres:
		if q.postgres != "" {
			return q.postgres
		}
		return q.any
	case cockroachdb:
		if q.cockroachdb != "" {
			return q.cockroachdb
		}
		return q.any

	case mysql:
		if q.mysql != "" {
			return q.mysql
		}
		return q.any
	case tidb:
		if q.tidb != "" {
			return q.tidb
		}
		return q.any
	default:
		panic("Unsupported database type")
	}
}

// queryFromTemplate is for easily creating queries that can run against multiple database backends
func queryFromTemplate(tmpl string) string {
	buff := bytes.NewBuffer([]byte{})
	paramCount := 0

	funcs := template.FuncMap{
		"?": func() string {
			paramCount++
			switch dbType {
			case postgres:
				return "$" + strconv.Itoa(paramCount)
			default:
				return "?"
			}
		},
		"offsetLimit": func(offset, limit int) string {
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
	}

	err := template.Must(template.New("").Funcs(funcs).Parse(tmpl)).Execute(buff, nil)
	if err != nil {
		panic(fmt.Sprintf("Error build query template: %s", err))
	}

	return buff.String()
}
