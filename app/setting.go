// Copyright (c) 2017 Townsourced Inc.

package app

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
	"github.com/pkg/errors"
)

// Setting is a defaulted value that changes how LexLibrary functions
type Setting struct {
	Key         string
	Description string
	Value       interface{}
	Options     []interface{}
	Category    string
}

// ErrSettingNotFound is returned when a setting can't be found for the given key
var ErrSettingNotFound = NotFound("No setting can be found for the given key")

// ErrSettingInvalidValue is returned when a setting is being set to a value that is invalid for the setting
var ErrSettingInvalidValue = NotFound("The setting cannot be set to this value")

var sqlSettingsGet = data.NewQuery("select key, value from settings")
var sqlSettingGet = data.NewQuery(`select value from settings where key = {{arg "key"}}`)
var sqlSettingUpdate = data.NewQuery(`update settings set value = {{arg "value"}} where key = {{arg "key"}}`)
var sqlSettingInsert = data.NewQuery(`
	insert into settings (key, description, value) values ({{arg "key"}}, {{arg "description"}}, {{arg "value"}})`)

// Settings returns all of the settings in Lex Library.  If a setting is not set in the database
// the default for that setting is returned
func Settings() ([]Setting, error) {
	settings := make([]Setting, len(settingDefaults))

	copy(settings, settingDefaults)

	rows, err := sqlSettingsGet.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var value string

		err = rows.Scan(&key, &value)
		if err != nil {
			return nil, err
		}

		for i := range settings {
			if settings[i].Key == key {
				err = settings[i].setValue(value)
				if err != nil {
					return nil, err
				}
				break
			}
		}

	}
	return settings, nil
}

// SettingGet will look for a setting that has the passed in key
func SettingGet(key string) (Setting, error) {
	var strValue string
	setting, err := SettingDefault(key)
	if err != nil {
		return Setting{}, err
	}

	err = sqlSettingGet.QueryRow(sql.Named("key", key)).Scan(&strValue)
	if err == sql.ErrNoRows {
		// nothing in the DB return the default setting value
		return setting, nil
	}
	if err != nil {
		// an error occurred retriving the setting from the DB
		// return the errror and the default, and let the consumer decide what to do
		return setting, errors.Wrapf(err, "Error getting setting %s", key)
	}

	err = setting.setValue(strValue)

	return setting, err
}

// SettingMust returns a setting.  If the setting does not exist it will panic
// Meant as a shortcut for setting lookups by the application
// SettingVal("AllowPublic").Bool()
func SettingMust(key string) Setting {
	setting, err := SettingGet(key)
	if err != nil {
		if err == ErrSettingNotFound {
			panic(fmt.Sprintf("Setting %s does not exist", key))
		}
		// if there is an error retriving the setting, log the error and return the default
		LogError(errors.Wrapf(err, "Error getting setting value for  %s, using default", key))
	}

	return setting
}

// SettingSet updates the value of the passed in setting to the passed in value
func SettingSet(key string, value interface{}) error {
	setting, err := SettingDefault(key)
	if err == ErrSettingNotFound {
		return err
	}

	if !setting.canSet(value) {
		return ErrSettingInvalidValue
	}

	strValue := ""
	switch value := value.(type) {
	case int:
		strValue = strconv.Itoa(value)
	case string:
		strValue = value
	case bool:
		strValue = strconv.FormatBool(value)
	case time.Duration:
		strValue = value.String()
	default:
		return ErrSettingInvalidValue
	}

	var tmp = ""
	err = sqlSettingGet.QueryRow(sql.Named("key", key)).Scan(&tmp)
	if err == sql.ErrNoRows {
		_, err := sqlSettingInsert.Exec(
			sql.Named("key", key),
			sql.Named("description", setting.Description),
			sql.Named("value", strValue))
		if err != nil {
			return errors.Wrapf(err, "Error inserting setting %s", key)
		}
		return nil
	}
	if err != nil {
		return err
	}

	_, err = sqlSettingUpdate.Exec(sql.Named("key", key), sql.Named("value", value))
	if err != nil {
		return errors.Wrapf(err, "Error updating setting %s", key)
	}
	return nil
}

// SettingDefault returns the default setting for the given setting key
func SettingDefault(key string) (Setting, error) {
	for i := range settingDefaults {
		if settingDefaults[i].Key == key {
			return settingDefaults[i], nil
		}
	}
	return Setting{}, ErrSettingNotFound
}

func (s *Setting) setValue(tableValue string) error {
	var value interface{}
	var err error

	switch s.Value.(type) {
	case int:
		value, err = strconv.Atoi(tableValue)
	case string:
		value = tableValue
	case bool:
		value, err = strconv.ParseBool(tableValue)
	case time.Duration:
		value, err = time.ParseDuration(tableValue)
	default:
		return errors.Errorf("Invalid setting type %T for setting %s", s.Value, s.Key)
	}

	if err != nil {
		// if a setting is stored in an un-parsable format in the DB, then update the db to the default value
		// in a a proper format and log that it occurred

		LogError(errors.Wrapf(err,
			"The value of setting %s in the database is in an invalid format (%s). Updating to default value", s.Key,
			tableValue))

		err = SettingSet(s.Key, s.Value)
		if err != nil {
			return errors.Wrapf(err, "Error setting default value for %s", s.Key)
		}
		return nil
	}

	s.Value = value
	return nil
}

// String returns the string value of the setting
// will panic if setting is not of type string
func (s *Setting) String() string {
	value, ok := s.Value.(string)
	if !ok {
		panic(fmt.Sprintf("Setting %s is not of type int", s.Key))
	}
	return value
}

// Int returns the int value of the setting
// will panic if setting is not of type int
func (s *Setting) Int() int {
	value, ok := s.Value.(int)
	if !ok {
		panic(fmt.Sprintf("Setting %s is not of type int", s.Key))
	}
	return value
}

// Bool returns the int value of the setting
// will panic if setting is not of type bool
func (s *Setting) Bool() bool {
	value, ok := s.Value.(bool)
	if !ok {
		panic(fmt.Sprintf("Setting %s is not of type bool", s.Key))
	}
	return value
}

// Duration returns the Duration value of the setting
// will panic if setting is not of type Duration
func (s *Setting) Duration() time.Duration {
	value, ok := s.Value.(time.Duration)
	if !ok {
		panic(fmt.Sprintf("Setting %s is not of type Duration", s.Key))
	}
	return value
}

// canSet tests if the passed in value can be set for the given setting
func (s *Setting) canSet(value interface{}) bool {
	if !reflect.TypeOf(value).AssignableTo(reflect.TypeOf(s.Value)) {
		return false
	}

	found := len(s.Options) == 0 //don't check options if there are none

optionsLoop:
	for i := range s.Options {
		switch value := value.(type) {
		case int:
			if value == s.Options[i].(int) {
				found = true
				break optionsLoop
			}
		case string:
			if value == s.Options[i].(string) {
				found = true
				break optionsLoop
			}
		case bool:
			if value == s.Options[i].(bool) {
				found = true
				break optionsLoop
			}
		case time.Duration:
			if value == s.Options[i].(time.Duration) {
				found = true
				break optionsLoop
			}
		default:
			found = false
			break optionsLoop
		}
	}

	return found
}
