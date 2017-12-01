// Copyright (c) 2017 Townsourced Inc.

package app_test

import (
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
)

func TestSetting(t *testing.T) {
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

}
