// Copyright (c) 2017-2018 Townsourced Inc.

package data

import (
	"database/sql/driver"

	"github.com/rs/xid"
)

var nilID ID

// ID is a globally unique ID
type ID xid.ID

// Scan implements the sql.Scanner interface.
func (i *ID) Scan(src interface{}) error {
	if src == nil {
		*i = nilID
		return nil
	}
	return (*xid.ID)(i).Scan(src)
}

// Value implements the driver.Valuer interface.
func (i ID) Value() (driver.Value, error) {
	return xid.ID(i).Value()
}

// NewID returns a new unique ID
func NewID() ID {
	return ID(xid.New())
}

// IDFromString returns an ID from the string value, errors if invalid
func IDFromString(val string) (ID, error) {
	id, err := xid.FromString(val)
	if err != nil {
		return nilID, err
	}

	return ID(id), nil
}

// IsNil returns whether or not the id is empty / nil
func (i ID) IsNil() bool {
	return i == nilID
}

func (i ID) String() string {
	return xid.ID(i).String()
}

// MarshalText implements encoding/text TextMarshaler interface
func (i ID) MarshalText() ([]byte, error) {
	return xid.ID(i).MarshalText()
}

// UnmarshalText implements encoding/text TextUnmarshaler interface
func (i *ID) UnmarshalText(text []byte) error {
	return (*xid.ID)(i).UnmarshalText(text)
}
