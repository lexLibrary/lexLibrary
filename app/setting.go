// Copyright (c) 2017 Townsourced Inc.

package app

import (
	"github.com/lexLibrary/lexLibrary/data"
)

// Setting is a defaulted value that changes how LexLibrary functions
type Setting struct {
	Key         string
	Description string
	Value       interface{}
}

// ErrSettingNotFound is returned when a setting can't be found for the given key
var ErrSettingNotFound = NotFound("No setting can be found for the given key")

var sqlSettingsGet = data.NewQuery("select key, value from settings")

// func Settings() ([]Setting, error) {

// }

// func SettingGet(key string) (*Setting, error) {

// }

// func SettingSet(key string, value interface{}) error {

// }

// SettingDefault returns the default setting for the given setting key
func SettingDefault(key string) (interface{}, error) {
	for i := range settingDefaults {
		if settingDefaults[i].Key == key {
			return settingDefaults[i].Value, nil
		}
	}
	return nil, ErrSettingNotFound
}

var settingDefaults = []Setting{
	Setting{
		Key:         "AllowPublic",
		Description: "Whether or not to allow documents to be published that are accessible without logging in to Lex Library",
		Value:       true,
	},
}
