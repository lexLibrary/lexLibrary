// Copyright (c) 2017-2018 Townsourced Inc.

package data

import (
	"database/sql/driver"

	"github.com/rs/xid"
)

// ID is a globally unique ID backed by xid
type ID struct {
	ID    xid.ID
	Valid bool
}

// NewID returns a new unique ID
func NewID() ID {
	return ID{
		Valid: true,
		ID:    xid.New(),
	}
}

// IDFromString returns an ID from the string value, errors if invalid
func IDFromString(val string) (ID, error) {
	id, err := xid.FromString(val)
	if err != nil {
		return ID{}, err
	}

	return ID{ID: id, Valid: true}, nil
}

// Scan implements the Scanner interface.
func (id *ID) Scan(value interface{}) error {
	id.Valid = true
	err := id.ID.Scan(value)
	if err != nil {
		id.Valid = false
	}
	return nil
}

// Value implements the driver Valuer interface.
func (id ID) Value() (driver.Value, error) {
	if !id.Valid {
		return nil, nil
	}
	return id.ID.Value()
}

// MarshalText implements encoding/text TextMarshaler interface
func (id ID) MarshalText() ([]byte, error) {
	if !id.Valid {
		return nil, nil
	}
	return id.ID.MarshalText()
}

// UnmarshalText implements encoding/text TextUnmarshaler interface
func (id *ID) UnmarshalText(text []byte) error {
	id.Valid = true
	err := id.ID.UnmarshalText(text)
	if err != nil {
		id.Valid = false
	}
	return nil
}

// String implements string interface
func (id ID) String() string {
	if !id.Valid {
		return ""
	}
	return id.ID.String()
}
