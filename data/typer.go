// Copyright (c) 2017-2018 Townsourced Inc.

package data

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/xid"
)

// Typer defines the interface needed to define a data type in Lex Library
// used to unify behaviors of database between the different Database backends and Go itself
type Typer interface {
	sql.Scanner
	driver.Valuer
}

// Text defines the text databtype, unlimited length, case sensitive unicode strings
type Text string

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

func (t *Text) Scan(src interface{}) error {
	var zero Text
	if src == nil {
		*t = zero
		return nil
	}
	s, ok := src.(string)
	if !ok {
		return errors.Errorf("Cannot scan value %v of type %T into Text", src, src)
	}
	*t = Text(s)

	return nil
}

func (t Text) Value() (driver.Value, error) {
	return t, nil
}

// VarChar  defines the varchar databtype, limited length, case sensitive unicode strings
type VarChar string

func varcharColumn(size int) string {
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

func (v *VarChar) Scan(src interface{}) error {
	var zero VarChar
	if src == nil {
		*v = zero
		return nil
	}
	s, ok := src.(string)
	if !ok {
		return errors.Errorf("Cannot scan value %v of type %T into VarChar", src, src)
	}
	*v = VarChar(s)

	return nil
}

func (v VarChar) Value() (driver.Value, error) {
	return v, nil
}

// ID is a globally unique ID
type ID xid.ID

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

func (i *ID) Scan(src interface{}) error {
	if src == nil {
		zero := xid.ID{}
		*i = ID(zero)
		return nil
	}
	return (*xid.ID)(i).Scan(src)
}

func (i ID) Value() (driver.Value, error) {
	return xid.ID(i).Value()
}

// DateTime dateTime defines a date and time with precision to miliseconds, stored in UTC
type DateTime time.Time

func datetimeColumn() string {
	// date + time + offset with precision to at least milliseconds
	switch dbType {
	case mysql, mariadb:
		return "DATETIME(5)"
	case sqlite:
		return "DATETIME"
	case postgres, cockroachdb:
		return "TIMESTAMP with time ZONE"
	case sqlserver:
		return "DATETIMEOFFSET"
	default:
		panic("Unsupported database type")
	}
}
func (d *DateTime) Scan(src interface{}) error {
	if src == nil {
		zero := DateTime{}
		*d = zero
		return nil
	}
	t, ok := src.(time.Time)
	if !ok {
		return errors.Errorf("Cannot scan value %v of type %T into DateTime", src, src)
	}

	*d = DateTime(t.UTC())

	return nil
}

func (d DateTime) Value() (driver.Value, error) {
	return d, nil
}

// Bool defines a non-null true / false value
type Bool bool

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

func (b *Bool) Scan(src interface{}) error {
	var zero Bool
	if src == nil {
		*b = zero
		return nil
	}
	s, ok := src.(bool)
	if !ok {
		return errors.Errorf("Cannot scan value %v of type %T into Bool", src, src)
	}
	*b = Bool(s)

	return nil
}

func (b Bool) Value() (driver.Value, error) {
	return b, nil
}

// Bytes defines variable length binary data
type Bytes []byte

// Int is a 64 bit integer
type Int int64
