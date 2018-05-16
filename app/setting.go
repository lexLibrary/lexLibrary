// Copyright (c) 2017-2018 Townsourced Inc.

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
	ID          string        `json:"id"`
	Description string        `json:"description"`
	Value       interface{}   `json:"value"`
	Options     []interface{} `json:"options"`
	validate    func(value interface{}) error
	triggers    []func(value interface{}) // function that gets called everytime this setting is set
}

// ErrSettingNotFound is returned when a setting can't be found for the given id
var ErrSettingNotFound = NotFound("No setting can be found for the given id")

// ErrSettingInvalidValue is returned when a setting is being set to a value that is invalid for the setting
var ErrSettingInvalidValue = NewFailure("The setting cannot be set to this value")

var (
	sqlSettingsGet   = data.NewQuery("select id, value from settings")
	sqlSettingGet    = data.NewQuery(`select value from settings where id = {{arg "id"}}`)
	sqlSettingUpdate = data.NewQuery(`update settings set value = {{arg "value"}} where id = {{arg "id"}}`)
	sqlSettingInsert = data.NewQuery(`insert into settings (id, description, value) 
		values ({{arg "id"}}, {{arg "description"}}, {{arg "value"}})`)
)

// Settings returns all of the settings in Lex Library.  If a setting is not set in the database
// the default for that setting is returned
func Settings(who *User) ([]Setting, error) {
	if who == nil || !who.Admin {
		return nil, Unauthorized("You must be an administrator to view settings")
	}
	return settings()
}
func settings() ([]Setting, error) {

	settings := make([]Setting, len(settingDefaults))

	copy(settings, settingDefaults)

	rows, err := sqlSettingsGet.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var value string

		err = rows.Scan(&id, &value)
		if err != nil {
			return nil, err
		}

		for i := range settings {
			if settings[i].ID == id {
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

func settingGet(id string) (Setting, error) {

	var strValue string
	setting, err := SettingDefault(id)
	if err != nil {
		return Setting{}, err
	}

	err = sqlSettingGet.QueryRow(data.Arg("id", id)).Scan(&strValue)
	if err == sql.ErrNoRows {
		// nothing in the DB return the default setting value
		return setting, nil
	}
	if err != nil {
		// an error occurred retrieving the setting from the DB
		// return the errror and the default, and let the consumer decide what to do
		return setting, errors.Wrapf(err, "Error getting setting %s", id)
	}

	err = setting.setValue(strValue)

	return setting, err
}

// SettingMust returns a setting.  If the setting does not exist it will panic
// Meant as a shortcut for setting lookups by the application
// SettingVal("AllowPublic").Bool()
func SettingMust(id string) Setting {
	setting, err := settingGet(id)
	if err != nil {
		if err == ErrSettingNotFound {
			panic(fmt.Sprintf("Setting %s does not exist", id))
		}
		// if there is an error retriving the setting, log the error and return the default
		LogError(errors.Wrapf(err, "Error getting setting value for  %s, using default", id))
	}

	return setting
}

// settingSetMultiple sets multiple settings in the same transaction
func settingSetMultiple(settings map[string]interface{}) error {
	return data.BeginTx(func(tx *sql.Tx) error {
		for id, val := range settings {
			err := settingSet(tx, id, val)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func settingSet(tx *sql.Tx, id string, value interface{}) error {
	setting, err := SettingDefault(id)
	if err == ErrSettingNotFound {
		return err
	}
	value, err = setting.convertValue(value)
	if err != nil {
		return err
	}

	if !setting.canSet(value) {
		return ErrSettingInvalidValue
	}

	if setting.validate != nil {
		err = setting.validate(value)
		if err != nil {
			return err
		}
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

	setting.Value = value
	var tmp = ""
	err = sqlSettingGet.Tx(tx).QueryRow(data.Arg("id", id)).Scan(&tmp)
	if err == sql.ErrNoRows {
		_, err := sqlSettingInsert.Tx(tx).Exec(
			data.Arg("id", id),
			data.Arg("description", setting.Description),
			data.Arg("value", strValue))
		if err != nil {
			return errors.Wrapf(err, "Error inserting setting %s", id)
		}
		setting.runTriggers()
		return nil
	}
	if err != nil {
		return err
	}

	_, err = sqlSettingUpdate.Tx(tx).Exec(data.Arg("id", id), data.Arg("value", value))
	if err != nil {
		return errors.Wrapf(err, "Error updating setting %s", id)
	}
	setting.runTriggers()
	return nil
}

// SettingDefault returns the default setting for the given setting id
func SettingDefault(id string) (Setting, error) {
	for i := range settingDefaults {
		if settingDefaults[i].ID == id {
			return settingDefaults[i], nil
		}
	}
	return Setting{}, ErrSettingNotFound
}

// convertValue returns the passed in value as the type needed for the setting
func (s *Setting) convertValue(value interface{}) (interface{}, error) {
	switch s.Value.(type) {
	case int:
		switch value := value.(type) {
		case float32:
			return int(value), nil
		case float64:
			return int(value), nil
		case int32:
			return int(value), nil
		case int64:
			return int(value), nil
		case int:
			return int(value), nil
		case string:
			i, err := strconv.Atoi(value)
			if err != nil {
				return nil, ErrSettingInvalidValue
			}
			return i, nil
		default:
			return nil, ErrSettingInvalidValue
		}
	case string:
		if value, ok := value.(string); ok {
			return value, nil
		}
	case bool:
		if value, ok := value.(bool); ok {
			return value, nil
		}
	case time.Duration:
		switch value := value.(type) {
		case float32:
			return time.Duration(value), nil
		case float64:
			return time.Duration(value), nil
		case int32:
			return time.Duration(value), nil
		case int64:
			return time.Duration(value), nil
		case int:
			return time.Duration(value), nil
		case string:
			i, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			return time.Duration(i), nil
		default:
			return nil, ErrSettingInvalidValue
		}
	default:
		return nil, ErrSettingInvalidValue
	}
	return nil, ErrSettingInvalidValue
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
		return errors.Errorf("Invalid setting type %T for setting %s", s.Value, s.ID)
	}

	if err != nil {
		// if a setting is stored in an un-parsable format in the DB, then update the db to the default value
		// in a a proper format and log that it occurred

		LogError(errors.Wrapf(err,
			"The value of setting %s in the database is in an invalid format (%s). Updating to default value", s.ID,
			tableValue))

		err = settingSet(nil, s.ID, s.Value)
		if err != nil {
			return errors.Wrapf(err, "Error setting default value for %s", s.ID)
		}
		return nil
	}

	s.Value = value
	return nil
}

// SettingTrigger registers a function to trigger when a setting value is updated, gets triggered once automatically
// when LexLibrary loads
func SettingTrigger(id string, triggerFunc func(value interface{})) {
	for i := range settingDefaults {
		if settingDefaults[i].ID == id {
			settingDefaults[i].triggers = append(settingDefaults[i].triggers, triggerFunc)
			return
		}
	}
	panic(fmt.Sprintf("Setting %s does not exist", id))
}

func settingTriggerInit() error {
	settings, err := settings()
	if err != nil {
		return err
	}

	for _, s := range settings {
		s.runTriggers()
	}
	return nil
}

// String returns the string value of the setting
// will panic if setting is not of type string
func (s Setting) String() string {
	value, ok := s.Value.(string)
	if !ok {
		panic(fmt.Sprintf("Setting %s is not of type int", s.ID))
	}
	return value
}

// Int returns the int value of the setting
// will panic if setting is not of type int
func (s Setting) Int() int {
	value, ok := s.Value.(int)
	if !ok {
		panic(fmt.Sprintf("Setting %s is not of type int", s.ID))
	}
	return value
}

// Bool returns the int value of the setting
// will panic if setting is not of type bool
func (s Setting) Bool() bool {
	value, ok := s.Value.(bool)
	if !ok {
		panic(fmt.Sprintf("Setting %s is not of type bool", s.ID))
	}
	return value
}

// Duration returns the Duration value of the setting
// will panic if setting is not of type Duration
func (s Setting) Duration() time.Duration {
	value, ok := s.Value.(time.Duration)
	if !ok {
		panic(fmt.Sprintf("Setting %s is not of type Duration", s.ID))
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

func (s *Setting) runTriggers() {
	for _, f := range s.triggers {
		f(s.Value)
	}
}
