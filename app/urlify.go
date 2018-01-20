// Copyright (c) 2017 Townsourced Inc.

package app

import (
	"unicode"
)

// for validating and making usernames or
// anything that needs to be url safe, case insensitive and user readable
type urlify string

// test if is valid urlForm ignoring case
func (u urlify) is() bool {
	if u == "" {
		return false
	}
	for _, c := range u {
		if !u.validRune(c) {
			return false
		}
	}
	return true
}

func (u urlify) validRune(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsNumber(c) || c == '-'
}

// func (u urlify) make() string {
// 	for i, c := range u {
// 		if !u.validRune(c) {
// 			//TODO: benchmark vs buffer
// 			u = u[:i] + "-" + u[i+1:]
// 		}
// 	}

// 	return strings.ToLower(string(u))
// }
