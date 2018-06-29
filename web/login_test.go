// Copyright (c) 2017-2018 Townsourced Inc.
package web_test

import (
	"testing"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
	"github.com/timshannon/sequence"
)

func TestLogin(t *testing.T) {
	uri := *llURL
	uri.Path = "login"

	data.ResetDB(t)

	username := "testusername"
	password := "testWithAPrettyGoodP@ssword"
	user, err := app.FirstRunSetup(username, password)
	if err != nil {
		t.Fatalf("Error setting up admin user: %s", err)
	}
	admin, err := user.Admin()
	if err != nil {
		t.Fatal(err)
	}

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
		Find(".has-error > .form-input-hint").Count(0).
		Find(".card-footer").Visible().
		Find("#inputUsername").SendKeys("badusername").
		Find("#inputPassword").SendKeys("badpassword").
		Find(".btn.btn-primary.btn-block").Click().
		Find(".has-error > .form-input-hint").Visible().
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
		Find(".btn.btn-primary.btn-block").Click().
		Find(".has-error > .form-input-hint").Count(0).
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
		Find(".btn.btn-primary.btn-block").Click().
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
		Exec(data.Arg("expires", time.Now().AddDate(0, 0, 6)), data.Arg("username", username))
	if err != nil {
		t.Fatalf("Error expiring password: %s", err)
	}

	err = sequence.Start(driver).
		Get(uri.String()).
		Find("#inputUsername").SendKeys(username).
		Find("#inputPassword").SendKeys(password).
		Find(".btn.btn-primary.btn-block").Click().
		Find(".has-error > .form-input-hint").Count(0).
		Find(".modal").Count(1).
		Find(".modal-overlay").Count(1).
		Find(".modal-container").Count(1).
		Find(".modal-footer > .btn").Any().Text().Contains("Skip").
		Find(".modal-footer > .btn.btn-primary").Any().Text().Contains("Submit").
		End()
	if err != nil {
		t.Fatalf("Testing expiring password failed: %s", err)
	}

	// expire password completely
	_, err = data.NewQuery(`update users set password_expiration = {{arg "expires"}}
			where username = {{arg "username"}}`).
		Exec(data.Arg("expires", time.Now()), data.Arg("username", username))
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
		Find(".btn.btn-primary.btn-block").Click().
		Find(".has-error > .form-input-hint").Count(0).
		Find(".modal").Count(1).
		Find(".modal-overlay").Count(1).
		Find(".modal-container").Count(1).
		Find(".modal-footer > .btn").Count(1).Text().Contains("Submit").Click().
		Find(".has-error > .form-input-hint").Text().Contains("You must provide a new password").
		Find("#inputNewPassword").SendKeys(password + "new").
		Find(".modal-footer > .btn").Click().
		Find(".has-error > .form-input-hint").Text().Contains("Passwords do not match").
		Find("#inputPassword2").SendKeys(password + "new").
		Find(".modal-footer > .btn").Click().
		Find(".has-error > .form-input-hint").Count(0).And().
		URL().Path("/").Eventually().
		End()
	if err != nil {
		t.Fatalf("Testing expiring password failed: %s", err)
	}

}
