// Copyright (c) 2017-2018 Townsourced Inc.
package web_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
	"github.com/timshannon/sequence"
)

func TestLogin(t *testing.T) {
	uri := *llURL
	uri.Path = "login"

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
	user, err := app.FirstRunSetup(username, password)
	if err != nil {
		t.Fatalf("Error setting up admin user: %s", err)
	}
	admin := user.AsAdmin()

	err = admin.SetSetting("AllowPublicSignups", true)
	if err != nil {
		t.Fatalf("Error allowing public signups for testing: %s", err)
	}

	err = driver.DeleteAllCookies()
	if err != nil {
		t.Fatalf("Error clearing all cookies for testing: %s", err)
	}

	// Invalid username and password
	err = newSequence().
		Get(uri.String()).
		Find("#login").Visible().
		Find("#inputUsername").Visible().
		Find(".help.is-danger").Count(0).
		Find(".card-footer").Visible().
		Find("#inputUsername").SendKeys("badusername").
		Find("#inputPassword").SendKeys("badpassword").
		Find(".button.is-primary.is-block").Click().
		Find(".help.is-danger").Visible().
		End()

	if err != nil {
		t.Fatalf("Testing Login Page failed: %s", err)
	}

	// Disabled Public Signups
	err = admin.SetSetting("AllowPublicSignups", false)
	if err != nil {
		t.Fatalf("Error blocking public signups for testing: %s", err)
	}

	err = newSequence().
		Refresh().
		Find(".card-footer").Count(0).
		Find("#inputUsername").SendKeys(username).
		Find("#inputPassword").SendKeys(password).
		Find(".button.is-primary.is-block").Click().
		Find(".help.is-danger").Count(0).
		End()

	if err != nil {
		t.Fatalf("Testing Login Page failed: %s", err)
	}

	// Page redirect on login
	err = driver.DeleteAllCookies()
	if err != nil {
		t.Fatalf("Error clearing all cookies for testing: %s", err)
	}

	testPath := "/testpath"
	err = newSequence().
		Get(uri.String() + "?return=" + testPath).
		Find("#inputUsername").SendKeys(username).
		Find("#inputPassword").SendKeys(password).
		Find(".button.is-primary.is-block").Click().
		And().
		URL().Path(testPath).Eventually().
		End()

	if err != nil {
		t.Fatalf("Testing Login Page failed: %s", err)
	}

	// Expire Password
	err = driver.DeleteAllCookies()
	if err != nil {
		t.Fatalf("Error clearing all cookies for testing: %s", err)
	}

	// expire password soon
	_, err = data.NewQuery(`update users set password_expiration = {{arg "expires"}}
			where username = {{arg "username"}}`).
		Exec(sql.Named("expires", time.Now().AddDate(0, 0, 6)), sql.Named("username", username))
	if err != nil {
		t.Fatalf("Error expiring password: %s", err)
	}

	err = sequence.Start(driver).
		Get(uri.String()).
		Find("#inputUsername").SendKeys(username).
		Find("#inputPassword").SendKeys(password).
		Find(".button.is-primary.is-block").Click().
		Find(".help.is-danger").Count(0).
		Find(".modal").Count(1).
		Find(".modal-background").Count(1).
		Find(".modal-card").Count(1).
		Find(".modal-card-foot > button").Any().Text().Contains("Skip").
		Find(".modal-card-foot > button").Any().Text().Contains("Submit").
		End()
	if err != nil {
		t.Fatalf("Testing expiring password failed: %s", err)
	}

	// expire password completely
	_, err = data.NewQuery(`update users set password_expiration = {{arg "expires"}}
			where username = {{arg "username"}}`).
		Exec(sql.Named("expires", time.Now()), sql.Named("username", username))
	if err != nil {
		t.Fatalf("Error expiring password: %s", err)
	}
	err = driver.DeleteAllCookies()
	if err != nil {
		t.Fatalf("Error clearing all cookies for testing: %s", err)
	}

	err = sequence.Start(driver).
		Get(uri.String()).
		Find("#inputUsername").SendKeys(username).
		Find("#inputPassword").SendKeys(password).
		Find(".button.is-primary.is-block").Click().
		Find(".help.is-danger").Count(0).
		Find(".modal").Count(1).
		Find(".modal-background").Count(1).
		Find(".modal-card").Count(1).
		Find(".modal-card-foot > button").Count(1).Text().Contains("Submit").Click().
		Find(".help.is-danger").Text().Contains("You must provide a new password").
		Find("#inputNewPassword").SendKeys(password + "new").
		Find(".modal-card-foot > button").Click().
		Find(".help.is-danger").Text().Contains("Passwords do not match").
		Find("#inputPassword2").SendKeys(password + "new").
		Find(".modal-card-foot > button").Click().
		Find(".help.is-danger").Count(0).And().
		URL().Path("/").
		End()
	if err != nil {
		t.Fatalf("Testing expiring password failed: %s", err)
	}

}
