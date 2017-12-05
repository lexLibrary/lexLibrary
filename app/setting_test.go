// Copyright (c) 2017 Townsourced Inc.

package app_test

import (
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
)

func TestSetting(t *testing.T) {
	_, err := data.NewQuery("delete from settings").Exec()
	if err != nil {
		t.Fatalf("Error emptying settings table before running tests: %s", err)
	}

	t.Run("Default", func(t *testing.T) {
		setting, err := app.SettingDefault("AllowPublic")
		if err != nil {
			t.Fatalf("Error getting setting default")
		}

		b, ok := setting.Value.(bool)
		if !ok {
			t.Fatalf("AllowPublic is not the correct type, expected bool, got %t", setting.Value)
		}

		if !b {
			t.Fatalf("AllowPublic setting isn't defaulted to true")
		}
	})
	t.Run("Default with invalid key", func(t *testing.T) {
		_, err := app.SettingDefault("badKey")
		if err == nil {
			t.Fatalf("No error returned from a bad default setting key")
		}

		if err != app.ErrSettingNotFound {
			t.Fatalf("Invalid error returned for Not Found key. Expected %v, got %v", app.ErrSettingNotFound, err)
		}
	})

	t.Run("Get", func(t *testing.T) {
		key := "AllowPublic"
		s, err := app.SettingGet(key)
		if err != nil {
			t.Fatalf("Error getting setting: %s", err)
		}
		if s.Key != key {
			t.Fatalf("Invalid Key returned. Expected %s got %s", key, s.Key)
		}

		d, err := app.SettingDefault(key)
		if err != nil {
			t.Fatalf("Error getting setting default: %s", err)
		}

		if d.Bool() != s.Bool() {
			t.Fatalf("Get did not return the default setting for an unset setting.  Expected %v, Got %v",
				d.Value, s.Value)
		}
	})

	t.Run("Get with Invalid key", func(t *testing.T) {
		_, err := app.SettingGet("badKey")
		if err == nil {
			t.Fatalf("No error returned from a bad get setting key")
		}

		if err != app.ErrSettingNotFound {
			t.Fatalf("Invalid error returned for Not Found key. Expected %v, got %v", app.ErrSettingNotFound, err)
		}
	})

	t.Run("Set", func(t *testing.T) {
		key := "AllowPublic"
		d, err := app.SettingDefault(key)
		if err != nil {
			t.Fatalf("Error getting setting default: %s", err)
		}

		err = app.SettingSet(key, !d.Bool())
		if err != nil {
			t.Fatalf("Error setting value: %s", err)
		}

		s, err := app.SettingGet(key)
		if err != nil {
			t.Fatalf("Error getting setting: %s", err)
		}

		if s.Bool() == d.Bool() {
			t.Fatalf("Setting value was not set.  Expected %v, got %v", !d.Bool(), s.Bool())
		}

		app.SettingSet(key, d.Bool())
		if err != nil {
			t.Fatalf("Error updating setting value: %s", err)
		}

		s, err = app.SettingGet(key)
		if err != nil {
			t.Fatalf("Error getting setting: %s", err)
		}

		if s.Bool() != d.Bool() {
			t.Fatalf("Setting value was not set.  Expected %v, got %v", d.Bool(), s.Bool())
		}

	})
	t.Run("Set with Invalid key", func(t *testing.T) {
		err := app.SettingSet("badKey", "badValue")
		if err == nil {
			t.Fatalf("No error returned from a bad Set setting key")
		}

		if err != app.ErrSettingNotFound {
			t.Fatalf("Invalid error returned for Not Found key. Expected %v, got %v", app.ErrSettingNotFound, err)
		}
	})
	t.Run("Set with Invalid Type", func(t *testing.T) {
		err := app.SettingSet("AllowPublic", "badValue")
		if err == nil {
			t.Fatalf("No error returned from a bad Set setting key")
		}

		if err != app.ErrSettingInvalidType {
			t.Fatalf("Invalid error returned for Not Found key. Expected %v, got %v", app.ErrSettingNotFound, err)
		}
	})

	t.Run("Bad setting format in database", func(t *testing.T) {
		_, err := data.NewQuery("update settings set value = 'garbage' where key = 'AllowPublic'").Exec()
		if err != nil {
			t.Fatalf("Error setting bad value in database: %s", err)
		}

		key := "AllowPublic"

		d, err := app.SettingDefault(key)
		if err != nil {
			t.Fatalf("Error getting default: %s", err)
		}

		s, err := app.SettingGet(key)
		if err != nil {
			t.Fatalf("Error getting setting: %s", err)
		}

		if s.Bool() != d.Bool() {
			t.Fatalf("Bad value wasn't updated to default value. Expected %v, got %v", d.Value, s.Value)
		}

	})

}
