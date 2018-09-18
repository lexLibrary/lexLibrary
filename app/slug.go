// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"bytes"
	"unicode"
)

func isSlugRune(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsNumber(c) || c == '-'
}

func isSlug(val string) bool {
	if val == "" {
		return false
	}
	for _, c := range val {
		if !isSlugRune(c) {
			return false
		}
	}
	return true
}

func makeSlug(val string) string {
	var b bytes.Buffer
	skip := false

	for _, c := range val {
		if isSlugRune(c) {
			b.WriteRune(unicode.ToLower(c))
			skip = false
		} else {
			if !skip {
				b.WriteRune('-')
				skip = true // don't write sequential '-' in a row
			}
		}
	}

	return b.String()
}
