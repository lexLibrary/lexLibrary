// Copyright (c) 2017-2018 Townsourced Inc.

package data

import (
	"github.com/rs/xid"
)

// NewID returns a new unique ID
func NewID() ID {
	return ID(xid.New())
}

// IDFromString returns an ID from the string value, errors if invalid
func IDFromString(val string) (ID, error) {
	id, err := xid.FromString(val)
	if err != nil {
		return ID(xid.ID{}), err
	}

	return ID(id), nil
}
