// Copyright (c) 2017-2018 Townsourced Inc.
package browser

import (
	"path"
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/browser/sequence"
	"github.com/lexLibrary/lexLibrary/data"
)

func TestLogin(t *testing.T) {
	username := "testusername"
	password := "testWithAPrettyGoodP@ssword"
	_, err := data.NewQuery("delete from users").Exec()
	if err != nil {
		t.Fatalf("Error emptying users table before running tests: %s", err)
	}
	_, err = data.NewQuery("delete from settings").Exec()
	if err != nil {
		t.Fatalf("Error emptying settings table before running tests: %s", err)
	}
	admin, err := app.FirstRunSetup(username, password)
	if err != nil {
		t.Fatalf("Error setting up admin user: %s", err)
	}

	err = app.SettingSet(admin, "AllowPublicSignups", true)
	if err != nil {
		t.Fatalf("Error allowing public signups for testing: %s", err)
	}

	err = driver.DeleteAllCookies()
	if err != nil {
		t.Fatalf("Error clearing all cookies for testing: %s", err)
	}

	err = sequence.Start(driver).
		Get(path.Join(uri, "/login")).
		Find("#login").Visible().
		Find("#inputUsername").Visible().
		Find(".invalid-feedback").Count(0).
		Find(".card-footer").Visible().
		Find("#inputUsername").SendKeys("badusername").
		Find("#inputPassword").SendKeys("badpassword").
		Find(".btn.btn-primary.btn-block").Click().
		Find(".invalid-feedback").Visible().
		End()

	if err != nil {
		t.Fatalf("Testing Login Page failed: %s", err)
	}

	err = app.SettingSet(admin, "AllowPublicSignups", false)
	if err != nil {
		t.Fatalf("Error blocking public signups for testing: %s", err)
	}

	err = sequence.Start(driver).
		Refresh().
		Find(".card-footer").Count(0).
		Find("#inputUsername").SendKeys(username).
		Find("#inputPassword").SendKeys(password).
		Find(".btn.btn-primary.btn-block").Click().
		End()

	if err != nil {
		t.Fatalf("Testing Login Page failed: %s", err)
	}

}
