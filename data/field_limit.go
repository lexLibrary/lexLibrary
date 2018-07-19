// Copyright (c) 2017-2018 Townsourced Inc.

package data

import (
	"fmt"
	"strings"
)

type fieldLimit struct {
	min int
	max int
}

// field limits should be all lower case and follow the naming standard <type>.<field>
// NOTE: If you update these after release, you'll need to to write column update scripts in schema_version.go
var fieldLimits = map[string]fieldLimit{
	"user.name":         {0, 64},
	"user.username":     {3, 64},
	"group.name":        {1, 128},
	"document.tag":      {1, 64},
	"document.language": {1, 50}, //https://tools.ietf.org/html/bcp47#section-4.4.1 - should be more than enough
}

// FieldLimit returns the limit for specific field sizes in lex library.  This provides one single location
// that can be referenced  at every layer (data, app, web) so that they are validated consistently
func FieldLimit(field string) fieldLimit {
	limit, ok := fieldLimits[field]
	if !ok {
		panic(fmt.Sprintf("No field found in data/field_limit.go for field %s", field))
	}
	return limit
}

// Max is the max length of the given field
func (f fieldLimit) Max() int { return f.max }

// Min is the min length of the given field
func (f fieldLimit) Min() int { return f.min }

// Valid tests the passed in value against the field's min and max
func (f fieldLimit) Valid(value string) bool {
	if len(value) > f.Max() {
		return false
	}
	if len(value) < f.Min() {
		return false
	}
	return true
}

// FieldValidate returns an err if the value is not valid
func FieldValidate(field, value string) error {
	limit := FieldLimit(field)

	id := strings.SplitN(field, ".", 2)[1]
	value = strings.TrimSpace(value)

	if len(value) > limit.Max() {
		return fmt.Errorf("%s must be less than %d characters", id, limit.Max())
	}
	if len(value) < limit.Min() {
		if limit.Min() == 1 {
			return fmt.Errorf("%s is a required field", id)
		}
		return fmt.Errorf("%s must be greater than %d characters", id, limit.Min())
	}

	return nil
}
