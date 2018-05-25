// Copyright (c) 2017-2018 Townsourced Inc.

package data

import (
	"database/sql/driver"
	"fmt"
	"time"
)

func textColumn() string {
	switch dbType {
	case sqlite, postgres, cockroachdb, mysql, mariadb:
		return "TEXT"
	case sqlserver:
		return "nvarchar(max)"
	default:
		panic("Unsupported database type")
	}
}

func varcharColumn(fieldOrSize interface{}) string {
	var size int
	switch val := fieldOrSize.(type) {
	case string:
		size = FieldLimit(val).Max()
	case int:
		size = val
	default:
		panic("Unsupported varchar value")
	}

	// case sensitive unicode with size limits
	switch dbType {
	case postgres, cockroachdb, mysql, mariadb:
		return fmt.Sprintf("varchar(%d)", size)
	case sqlite:
		return "TEXT"
	case sqlserver:
		return fmt.Sprintf("nvarchar(%d)", size)
	default:
		panic("Unsupported database type")
	}
}

func idColumn() string {
	// case sensitive unicode with size limits
	switch dbType {
	case postgres, cockroachdb, mysql, mariadb:
		return "varchar(20)"
	case sqlite:
		return "TEXT"
	case sqlserver:
		return fmt.Sprintf("nvarchar(20)")
	default:
		panic("Unsupported database type")
	}
}

func datetimeColumn() string {
	// date + time stored in UTC with precision to at least milliseconds
	switch dbType {
	case mysql, mariadb:
		return "DATETIME(5)"
	case sqlite:
		return "DATETIME"
	case postgres, cockroachdb:
		return "TIMESTAMP"
	case sqlserver:
		return "DATETIME2"
	default:
		panic("Unsupported database type")
	}
}

func bytesColumn() string {
	// binary data with no size limits
	switch dbType {
	case sqlite:
		return "BLOB"
	case postgres:
		return "BYTEA"
	case cockroachdb:
		return "BYTES"
	case mysql, mariadb:
		return "BLOB"
	case sqlserver:
		return "VARBINARY(max)"
	default:
		panic("Unsupported database type")
	}
}

func boolColumn() string {
	switch dbType {
	case postgres, cockroachdb, mysql, mariadb:
		return "boolean"
	case sqlserver:
		return "bit"
	case sqlite:
		return "int"
	default:
		panic("Unsupported database type")
	}
}

func intColumn() string {
	// 64bit integers
	switch dbType {
	case sqlite:
		return "int"
	case postgres, mysql, mariadb, cockroachdb, sqlserver:
		return "bigint"
	default:
		panic("Unsupported database type")
	}
}

// NullTime represents a time.Time that may be null. NullTime implements the
// sql.Scanner interface so it can be used as a scan destination, similar to
// sql.NullString.
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Scan implements the Scanner interface.
func (nt *NullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	return nil
}

// Value implements the driver Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// MarshalJSON implements the JSON interface for NullTime
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return nt.Time.MarshalJSON()
}

// UnmarshalJSON implements the JSON interface for NullTime
func (nt *NullTime) UnmarshalJSON(data []byte) error {
	if data == nil {
		nt.Valid = false
	}
	return nt.Time.UnmarshalJSON(data)
}
