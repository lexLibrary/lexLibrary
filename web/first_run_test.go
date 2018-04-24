// Copyright (c) 2017-2018 Townsourced Inc.
package web_test

import (
	"errors"
	"strings"

	"github.com/tebeka/selenium"
)

func firstRun() error {
	err := reset()
	if err != nil {
		return err
	}

	err = newSequence().
		Get(llURL.String()).
		Find("#submit").Click().
		Find(".has-error > .form-input-hint").Text().Contains("A username is required").
		Find("#inputUsername").SendKeys("testusername").
		Find("#submit").Click().
		Find(".has-error > .form-input-hint").Text().Contains("A password is required").
		Find("#inputPassword").SendKeys("testWithAPrettyGoodP@ssword").
		Find("#submit").Click().
		Find(".has-error > .form-input-hint").Count(1).Text().Contains("Passwords do not match").
		Find("#inputPassword2").SendKeys("testWithAPrettyGoodP@ssword").
		Find("#submit").Click().
		Find(".has-error > .form-input-hint").Count(0).
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
		Find(".toast.toast-error").Count(0).
		End()
	if err != nil {
		return err
	}
	return nil
}
