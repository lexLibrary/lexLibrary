// Copyright (c) 2017-2018 Townsourced Inc.
package browser

import (
	"errors"
	"strings"
	"testing"

	"github.com/lexLibrary/lexLibrary/browser/sequence"
	"github.com/lexLibrary/lexLibrary/data"
	"github.com/tebeka/selenium"
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
		Get(uri).
		Find("#submit").Click().
		Find(".alert.alert-danger").Text().Contains("A username is required").
		Find("#inputUsername").SendKeys("testusername").
		Find("#submit").Click().
		Find(".alert.alert-danger").Text().Contains("A password is required").
		Find("#inputPassword").SendKeys("testWithAPrettyGoodP@ssword").
		Find("#submit").Click().
		Find(".alert.alert-danger").Count(0).
		Find(".invalid-feedback").Count(1).Text().Contains("Passwords do not match").
		Find("#inputPassword2").SendKeys("testWithAPrettyGoodP@ssword").
		Find("#submit").Click().
		Find(".alert.alert-danger").Count(0).
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
		Find(".alert.alert-danger").Count(0).
		End()
	if err != nil {
		t.Fatalf("Testing First run failed: %s", err)
	}

}
