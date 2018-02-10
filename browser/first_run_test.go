// Copyright (c) 2017-2018 Townsourced Inc.
package browser

import (
	"testing"

	"github.com/lexLibrary/lexLibrary/browser/sequence"
	"github.com/lexLibrary/lexLibrary/data"
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
		Find("#inputUsername").Submit().And().
		Find(".alert.alert-danger").Text().Contains("A username is required").And().
		Find("#inputUsername").SendKeys("testusername").Submit().And().
		Find(".alert.alert-danger").Text().Contains("A password is required").And().
		// Find("#inputPassword").SendKeys("testWithAPrettyGoodP@ssword").Submit().And().
		// Find(".alert.alert-danger").Count(0).And().
		// Find("#inputUsername").Count(0).And().
		// Find("#settings").Count(1).And().
		// Test("LL Cookie", func(d selenium.WebDriver) error {
		// 	c, err := d.GetCookie("lexlibrary")
		// 	if err != nil {
		// 		return err
		// 	}
		// 	if !strings.Contains(c.Value, "@") {
		// 		return errors.New("Invalid Lex Library Cookie")
		// 	}
		// 	return nil
		// }).
		// Find("#showAdvanced").Click().
		End()
	if err != nil {
		t.Fatal("Testing First run failed: ", err)
	}

}
