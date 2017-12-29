// Copyright (c) 2017 Townsourced Inc.

package app

import (
	"runtime"
)

type passworder interface {
	hash(password string) ([]byte, error)
	compare(password string, hash []byte) error // nil on success, err on failure
}

// ErrLogonFailure is when a user fails a login attempt
var ErrLogonFailure = NewFailure("Invalid user and / or password")

// index of the passwordVersions array is the version of the password set in users.password_version
var passwordVersions = []passworder{
	&argon{
		time:    4,
		memory:  32 * 1024,
		threads: uint8(runtime.GOMAXPROCS(-1)),
		keyLen:  32,
	},
}

// validatePassword validates if the passed in password
// meets the requirements for a good password in lex library
func validatePassword(password string) error {
	// check length
	// check against bad password list
	return nil
}
