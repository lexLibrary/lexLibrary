// Copyright (c) 2017-2018 Townsourced Inc.
package web_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
	"github.com/lexLibrary/sequence"
	"github.com/tebeka/selenium"
)

func TestSignup(t *testing.T) {
	uri := *llURL
	uri.Path = "signup"

	adminUsername := "admin"
	password := "testWithAPrettyGoodP@ssword"
	_, err := data.NewQuery("delete from users").Exec()
	if err != nil {
		t.Fatalf("Error emptying users table before running tests: %s", err)
	}
	_, err = data.NewQuery("delete from settings").Exec()
	if err != nil {
		t.Fatalf("Error emptying settings table before running tests: %s", err)
	}
	user, err := app.FirstRunSetup(adminUsername, password)
	if err != nil {
		t.Fatalf("Error setting up admin user: %s", err)
	}
	admin := user.AsAdmin()

	err = admin.SetSetting("AllowPublicSignups", false)
	if err != nil {
		t.Fatalf("Error allowing public signups for testing: %s", err)
	}

	err = driver.DeleteAllCookies()
	if err != nil {
		t.Fatalf("Error clearing all cookies for testing: %s", err)
	}

	err = sequence.Start(driver).
		Get(uri.String()).
		Title().Equals("Page Not Found - Lex Library").
		End()

	if err != nil {
		t.Fatalf("Testing Signup Page failed: %s", err)
	}

	err = admin.SetSetting("AllowPublicSignups", true)
	if err != nil {
		t.Fatalf("Error allowing public signups for testing: %s", err)
	}

	err = sequence.Start(driver).
		Get(uri.String()).
		Find("#submit").Click().
		Find(".help.is-danger").Count(2).Any().Text().Contains("A username is required").
		Find("#inputUsername").SendKeys(admin.User.Username).
		Find("#submit").Click().
		Find(".help.is-danger").Any().Text().Contains("This username is already taken").
		Find("#inputUsername").SendKeys("testusername").
		Find("#submit").Click().
		Find(".help.is-danger").Text().Contains("A password is required").
		Find("#inputPassword").SendKeys("bad").
		Find("#submit").Click().
		Find(".help.is-danger").Any().Text().Contains("The password must be at least").
		Find("#inputPassword").Clear().SendKeys("testWithAPrettyGoodP@ssword").
		Find("#submit").Click().
		Find(".help.is-danger").Text().Contains("Passwords do not match").
		Find("#inputPassword2").SendKeys("testWithAPrettyGoodP@ssword").
		Find("#submit").Click().
		Find(".help.is-danger").Count(0).
		And().
		Test("LL Cookie", func(d selenium.WebDriver) error {
			c, err := d.GetCookie("lexlibrary")
			if err != nil {
				return err
			}
			if !strings.Contains(c.Value, "@") {
				return errors.New("Invalid Lex Library Cookie")
			}
			return nil
		}).
		End()

	if err != nil {
		t.Fatalf("Testing Signup Page failed: %s", err)
	}

}
