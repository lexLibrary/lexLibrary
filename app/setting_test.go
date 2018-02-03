// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"reflect"
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
)

func TestSetting(t *testing.T) {
	_, err := data.NewQuery("delete from users").Exec()
	if err != nil {
		t.Fatalf("Error emptying users table before running tests: %s", err)
	}
	admin, err := app.FirstRunSetup("admin", "adminpassword")
	if err != nil {
		t.Fatalf("Error setting up admin user: %s", err)
	}

	reset := func() {
		t.Helper()
		_, err := data.NewQuery("delete from settings").Exec()
		if err != nil {
			t.Fatalf("Error emptying settings table before running tests: %s", err)
		}

	}

	t.Run("Default", func(t *testing.T) {
		reset()
		setting, err := app.SettingDefault("AllowPublicDocuments")
		if err != nil {
			t.Fatalf("Error getting setting default")
		}

		b, ok := setting.Value.(bool)
		if !ok {
			t.Fatalf("AllowPublicDocuments is not the correct type, expected bool, got %t", setting.Value)
		}

		if !b {
			t.Fatalf("AllowPublicDocuments setting isn't defaulted to true")
		}
	})
	t.Run("Default with invalid id", func(t *testing.T) {
		_, err := app.SettingDefault("badKey")
		if err == nil {
			t.Fatalf("No error returned from a bad default setting id")
		}

		if err != app.ErrSettingNotFound {
			t.Fatalf("Invalid error returned for Not Found id. Expected %v, got %v", app.ErrSettingNotFound, err)
		}
	})

	t.Run("Get", func(t *testing.T) {
		id := "AllowPublicDocuments"
		s, err := app.SettingGet(admin, id)
		if err != nil {
			t.Fatalf("Error getting setting: %s", err)
		}
		if s.ID != id {
			t.Fatalf("Invalid ID returned. Expected %s got %s", id, s.ID)
		}

		d, err := app.SettingDefault(id)
		if err != nil {
			t.Fatalf("Error getting setting default: %s", err)
		}

		if d.Bool() != s.Bool() {
			t.Fatalf("Get did not return the default setting for an unset setting.  Expected %v, Got %v",
				d.Value, s.Value)
		}
	})

	t.Run("Get with Invalid id", func(t *testing.T) {
		_, err := app.SettingGet(admin, "badKey")
		if err == nil {
			t.Fatalf("No error returned from a bad get setting id")
		}

		if err != app.ErrSettingNotFound {
			t.Fatalf("Invalid error returned for Not Found id. Expected %v, got %v", app.ErrSettingNotFound, err)
		}
	})

	t.Run("Set", func(t *testing.T) {
		id := "AllowPublicDocuments"
		d, err := app.SettingDefault(id)
		if err != nil {
			t.Fatalf("Error getting setting default: %s", err)
		}

		err = app.SettingSet(admin, id, !d.Bool())
		if err != nil {
			t.Fatalf("Error setting value: %s", err)
		}

		s, err := app.SettingGet(admin, id)
		if err != nil {
			t.Fatalf("Error getting setting: %s", err)
		}

		if s.Bool() == d.Bool() {
			t.Fatalf("Setting value was not set.  Expected %v, got %v", !d.Bool(), s.Bool())
		}

		err = app.SettingSet(admin, id, d.Bool())
		if err != nil {
			t.Fatalf("Error updating setting value: %s", err)
		}

		s, err = app.SettingGet(admin, id)
		if err != nil {
			t.Fatalf("Error getting setting: %s", err)
		}

		if s.Bool() != d.Bool() {
			t.Fatalf("Setting value was not set.  Expected %v, got %v", d.Bool(), s.Bool())
		}

	})
	t.Run("Set with Invalid id", func(t *testing.T) {
		err := app.SettingSet(admin, "badKey", "badValue")
		if err == nil {
			t.Fatalf("No error returned from a bad Set setting id")
		}

		if err != app.ErrSettingNotFound {
			t.Fatalf("Invalid error returned for Not Found id. Expected %v, got %v", app.ErrSettingNotFound, err)
		}
	})
	t.Run("Set with Invalid Type", func(t *testing.T) {
		err := app.SettingSet(admin, "AllowPublicDocuments", "badValue")
		if err == nil {
			t.Fatalf("No error returned from a bad Set setting id")
		}

		if err != app.ErrSettingInvalidValue {
			t.Fatalf("Invalid error returned for Not Found id. Expected %v, got %v", app.ErrSettingNotFound, err)
		}
	})

	t.Run("Bad setting format in database", func(t *testing.T) {
		_, err := data.NewQuery("update settings set value = 'garbage' where id = 'AllowPublicDocuments'").Exec()
		if err != nil {
			t.Fatalf("Error setting bad value in database: %s", err)
		}

		id := "AllowPublicDocuments"

		d, err := app.SettingDefault(id)
		if err != nil {
			t.Fatalf("Error getting default: %s", err)
		}

		s, err := app.SettingGet(admin, id)
		if err != nil {
			t.Fatalf("Error getting setting: %s", err)
		}

		if s.Bool() != d.Bool() {
			t.Fatalf("Bad value wasn't updated to default value. Expected %v, got %v", d.Value, s.Value)
		}

	})

	t.Run("Settings List", func(t *testing.T) {
		reset()

		settings, err := app.Settings(admin)
		if err != nil {
			t.Fatalf("Couldn't get list of all settings: %s", err)
		}

		for i := range settings {
			d, err := app.SettingDefault(settings[i].ID)
			if err != nil {
				t.Fatalf("Error getting default setting for %s: %s", settings[i].ID, err)
			}

			if !reflect.DeepEqual(d.Value, settings[i].Value) {
				t.Fatalf("Unset setting value was not defaulted.  Expected %v, Got %v", d.Value, settings[i].Value)
			}
		}

		id := "AllowPublicDocuments"
		d, err := app.SettingDefault(id)
		if err != nil {
			t.Fatalf("Error getting default setting for %s: %s", id, err)
		}

		err = app.SettingSet(admin, id, !d.Bool())
		if err != nil {
			t.Fatalf("Error setting setting: %s", err)
		}

		settings, err = app.Settings(admin)
		if err != nil {
			t.Fatalf("Couldn't get list of all settings: %s", err)
		}

		for i := range settings {
			if settings[i].ID == id {
				if settings[i].Bool() == d.Bool() {
					t.Fatalf("Settings value didn't update. Expected %v got %v", d.Bool(), settings[i].Bool())
					break
				}
			}
		}

	})

	t.Run("Must", func(t *testing.T) {
		reset()
		id := "AllowPublicDocuments"
		s := app.SettingMust(id)
		if s.ID != id {
			t.Fatalf("Invalid ID returned. Expected %s got %s", id, s.ID)
		}

		d, err := app.SettingDefault(id)
		if err != nil {
			t.Fatalf("Error getting setting default: %s", err)
		}

		if d.Bool() != s.Bool() {
			t.Fatalf("Must did not return the default setting for an unset setting.  Expected %v, Got %v",
				d.Value, s.Value)
		}
	})

	t.Run("Must with Invalid id", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("SettingMust did not panic with a bad id")
			}
		}()

		_ = app.SettingMust("badKey")
	})

	t.Run("Settings with options", func(t *testing.T) {
		reset()

		err := app.SettingSet(admin, "AuthenticationType", "bad value")
		if err == nil {
			t.Fatalf("Setting an an invalid value on a setting with options did not fail")
		}

		if err != app.ErrSettingInvalidValue {
			t.Fatalf("Invalid value on a setting with options did not return the correct error. Expected %s got %s",
				app.ErrSettingInvalidValue, err)
		}
	})

	t.Run("Setting with validate func", func(t *testing.T) {
		reset()

		err := app.SettingSet(admin, "PasswordMinLength", 3)
		if err == nil {
			t.Fatalf("Setting an an invalid value on a setting with a validate func did not fail")
		}
	})

	t.Run("Setting Triggers", func(t *testing.T) {
		reset()
		got := 0
		app.SettingTrigger("PasswordMinLength", func(value interface{}) {
			got = value.(int)
		})

		expected := 20
		err := app.SettingSet(admin, "PasswordMinLength", expected)
		if err != nil {
			t.Fatalf("Error setting setting value: %s", err)
		}
		if got != expected {
			t.Fatalf("Setting trigger did not run. Expected %d, got %d", expected, got)
		}

	})
}
