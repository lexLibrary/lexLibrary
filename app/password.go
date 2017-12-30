// Copyright (c) 2017 Townsourced Inc.

package app

import (
	"bytes"
	"runtime"
	"unicode"

	"github.com/lexLibrary/lexLibrary/files"
)

const passwordSpecialCharacters = ` !"#$%&'()*+,-./:;<=>?@[\]^_` + "`{|}~"

type passworder interface {
	hash(password string) ([]byte, error)
	compare(password string, hash []byte) error // nil on success, err on failure
}

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
	min := SettingMust("PasswordMinLength").Int()

	if len(password) < min {
		return NewFailure("The password must be at least %d characters long", min)
	}

	special := SettingMust("PasswordRequireSpecial").Bool()
	number := SettingMust("PasswordRequireNumber").Bool()
	mixed := SettingMust("PasswordRequireMixedCase").Bool()

	specialFound, numberFound, upperFound, lowerFound := false, false, false, false

	if special || number || mixed {
		for _, r := range password {
			if number && unicode.IsNumber(r) {
				numberFound = true
			}

			if special {
				for _, s := range passwordSpecialCharacters {
					if s == r {
						specialFound = true
						break
					}
				}
			}
			if mixed && unicode.IsUpper(r) {
				upperFound = true
			}
			if mixed && unicode.IsLower(r) {
				lowerFound = true
			}
		}

		if special && !specialFound {
			return NewFailure("The password must contain at least one special character")
		}
		if number && !numberFound {
			return NewFailure("The password must contain at least one number")
		}
		if mixed && (!upperFound || !lowerFound) {
			return NewFailure("The password must contain at least one upper case character")
		}
	}

	if SettingMust("BadPasswordCheck").Bool() {
		bad, err := files.Asset("app/bad_passwords.txt")
		if err != nil {
			return err
		}

		if bytes.Contains(bad, []byte(password)) {
			return NewFailure("Your password is commonly used.  Please choose a new, more unique password.")
		}
	}

	return nil
}
