// Copyright (c) 2017-2018 Townsourced Inc.
package web_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/lexLibrary/lexLibrary/data"
	"github.com/tebeka/selenium"
	"github.com/timshannon/sequence"
)

func TestFirstRun(t *testing.T) {
	_, err := data.NewQuery("delete from users").Exec()
	if err != nil {
		t.Fatalf("Error emptying users table before running tests: %s", err)
	}
	_, err = data.NewQuery("delete from settings").Exec()
	if err != nil {
		t.Fatalf("Error emptying settings table before running tests: %s", err)
	}

	err = sequence.Start(driver).
		Get(llURL.String()).
		Find("#submit").Click().
		Find(".help.is-danger").Text().Contains("A username is required").
		Find("#inputUsername").SendKeys("testusername").
		Find("#submit").Click().
		Find(".help.is-danger").Text().Contains("A password is required").
		Find("#inputPassword").SendKeys("testWithAPrettyGoodP@ssword").
		Find("#submit").Click().
		Find(".help.is-danger").Count(1).Text().Contains("Passwords do not match").
		Find("#inputPassword2").SendKeys("testWithAPrettyGoodP@ssword").
		Find("#submit").Click().
		Find(".help.is-danger").Count(0).
		Find("#inputUsername").Count(0).
		Find("#settings").Count(1).And().
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
		Find("#inputPrivate").Enabled().
		Find("#inputPublic").Enabled().
		Find("#allowPublicDocs").Count(0).
		Find("#allowPublicSignup").Count(0).
		Find("#showAdvanced").Click().
		Find("#allowPublicDocs").Count(1).
		Find("#allowPublicSignup").Count(1).
		Find("#inputPrivate").Disabled().
		Find("#inputPublic").Disabled().
		Find("#setSettings").Click().
		Find(".notification.is-danger").Count(0).
		End()
	if err != nil {
		t.Fatalf("Testing First run failed: %s", err)
	}

}
