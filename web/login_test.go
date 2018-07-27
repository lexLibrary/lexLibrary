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
	ok(t, err)

	admin, err := user.Admin()
	ok(t, err)

	err = admin.SetSetting("AllowPublicSignups", true)
	ok(t, err)

	ok(t, driver.DeleteAllCookies())

	// Invalid username and password
	newSequence().
		Get(uri.String()).
		Find("#login").Visible().
		Find("#inputUsername").Visible().
		Find(".toast-error").Count(0).
		Find(".card-footer").Visible().
		Find("#inputUsername").SendKeys("badusername").
		Find("#inputPassword").SendKeys("badpassword").
		Find(".btn.btn-primary.btn-block").Click().
		Find(".toast-error").Visible().
		Ok(t)

	// Disabled Public Signups
	ok(t, admin.SetSetting("AllowPublicSignups", false))

	newSequence().
		Refresh().
		Find(".card-footer").Count(0).
		Find("#inputUsername").SendKeys(username).
		Find("#inputPassword").SendKeys(password).
		Find(".btn.btn-primary.btn-block").Click().
		Find(".toast-error").Count(0).
		Ok(t)

	// Page redirect on login
	ok(t, driver.DeleteAllCookies())

	testPath := "/testpath"
	newSequence().
		Get(uri.String() + "?return=" + testPath).
		Find("#inputUsername").SendKeys(username).
		Find("#inputPassword").SendKeys(password).
		Find(".btn.btn-primary.btn-block").Click().
		And().
		URL().Path(testPath).Eventually().
		Ok(t)

	// Expire Password
	ok(t, driver.DeleteAllCookies())

	// expire password soon
	_, err = data.NewQuery(`update users set password_expiration = {{arg "expires"}}
			where username = {{arg "username"}}`).
		Exec(data.Arg("expires", time.Now().AddDate(0, 0, 6)), data.Arg("username", username))
	ok(t, err)

	sequence.Start(driver).
		Get(uri.String()).
		Find("#inputUsername").SendKeys(username).
		Find("#inputPassword").SendKeys(password).
		Find(".btn.btn-primary.btn-block").Click().
		Find(".toast-error").Count(0).
		Find(".modal").Count(1).
		Find(".modal-overlay").Count(1).
		Find(".modal-container").Count(1).
		Find(".modal-footer > .btn").Any().Text().Contains("Skip").
		Find(".modal-footer > .btn.btn-primary").Any().Text().Contains("Submit").
		Ok(t)

	// expire password completely
	_, err = data.NewQuery(`update users set password_expiration = {{arg "expires"}}
			where username = {{arg "username"}}`).
		Exec(data.Arg("expires", time.Now()), data.Arg("username", username))
	ok(t, err)
	ok(t, driver.DeleteAllCookies())

	sequence.Start(driver).
		Get(uri.String()).
		Find("#inputUsername").SendKeys(username).
		Find("#inputPassword").SendKeys(password).
		Find(".btn.btn-primary.btn-block").Click().
		Find(".toast-error").Count(0).
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
		Ok(t)
}
