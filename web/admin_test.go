// Copyright (c) 2017-2018 Townsourced Inc.

package web_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/tebeka/selenium"
	"github.com/timshannon/sequence"
)

func TestAdmin(t *testing.T) {
	uri := *llURL
	uri.Path = "admin"

	username := "testuser"
	password := "testpasswordThatisLongEnough"

	err := createUserAndLogin(username, password, true)
	if err != nil {
		t.Fatalf("Error setting up user for testing: %s", err)
	}

	seq := newSequence()

	t.Run("Overview", func(t *testing.T) {
		err = seq.Get(uri.String()).
			Find(".tab > .tab-item.active").Count(1).Text().Contains("Overview").
			Find(".admin-overview.table").Count(6).
			Find(".card > .card-header > .card-title").Any().Text().Contains("Instance Information").
			Find(".card > .card-header > .card-title").Any().Text().Contains("Data Usage").
			Find(".card > .card-header > .card-title").Any().Text().Contains("System Information").
			Find(".card > .card-header > .card-title").Any().Text().Contains("Runtime").
			Find(".card > .card-header > .card-title").Any().Text().Contains("Web Configuration").
			Find(".card > .card-header > .card-title").Any().Text().Contains("Data Configuration").
			Find(".tab > .tab-item > a[href='/admin']").Click().
			Find(".tab > .tab-item.active").Count(1).Text().Contains("Overview").
			End()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Logs", func(t *testing.T) {
		errMessage := "New Error message"
		errID := app.LogError(fmt.Errorf(errMessage))

		err = seq.
			Find(".tab > .tab-item > a[href='/admin/logs']").Click().
			Find(".tab > .tab-item.active").Count(1).Text().Contains("Logs").
			Find(".admin-logs > table.table > tbody > tr").Text().Contains(errMessage).
			Find(".admin-logs > table.table > tbody > tr > td > a.float-right").Text().Contains("View").Click().
			Find("h4").Text().Contains("Log Entry from").
			Find("h4 > small").Text().Contains(errID.String()).
			Find("section.admin-logs > p").Text().Contains(errMessage).
			End()
		if err != nil {
			t.Fatal(err)
		}

		// searching
		app.LogError(fmt.Errorf("other error"))
		err = seq.Back().Refresh().
			Find(".admin-logs > table.table > tbody > tr").Count(2).
			Find(".input-group > input.input[type='text']").SendKeys(strings.ToUpper(errMessage)).
			Find(".input-group > button.btn.btn-primary").Click().
			Find(".admin-logs > table.table > tbody > tr").Count(1).Text().Contains(errMessage).
			Find(".input-group > input.input[type='text']").SendKeys(errID.String()).
			Find(".input-group > button.btn.btn-primary").Click().
			Find("h4").Text().Contains("Log Entry from").
			Find("h4 > small").Text().Contains(errID.String()).
			Find("section.admin-logs > p").Text().Contains(errMessage).
			End()
		if err != nil {
			t.Fatal(err)
		}

		// pagination

		addPage := func() {
			// add one page of logs
			for i := 0; i < 30; i++ {
				app.LogError(fmt.Errorf("Error number: %d", i))
			}
		}

		testPagination := func(driver selenium.WebDriver, pageLinks []string) error {
			seq := sequence.Start(driver)
			for i := range pageLinks {
				seq = seq.Find(fmt.Sprintf("ul.pagination > li:nth-child(%d).page-item", i+1)).
					Text().Contains(pageLinks[i]).And()
			}
			return seq.End()
		}

		uri.Path = "/admin/logs"

		addPage()
		err = seq.Get(uri.String()).
			Find(".admin-logs > table.table > tbody > tr").Count(30).
			Find("ul.pagination > li.page-item.disabled").Text().Contains("Previous").
			Find("ul.pagination > li:nth-child(2).page-item.active").Text().Contains("1").And().
			Test("pagination", func(driver selenium.WebDriver) error {
				return testPagination(driver, []string{"Previous", "1", "2", "Next"})
			}).End()

		if err != nil {
			t.Fatal(err)
		}

		addPage()
		err = seq.Get(uri.String()).
			Find(".admin-logs > table.table > tbody > tr").Count(30).
			Find("ul.pagination > li.page-item.disabled").Text().Contains("Previous").
			Find("ul.pagination > li:nth-child(2).page-item.active").Text().Contains("1").And().
			Test("pagination", func(driver selenium.WebDriver) error {
				return testPagination(driver, []string{"Previous", "1", "2", "3", "Next"})
			}).End()

		if err != nil {
			t.Fatal(err)
		}

		addPage()
		err = seq.Get(uri.String()).
			Find(".admin-logs > table.table > tbody > tr").Count(30).
			Find("ul.pagination > li.page-item.disabled").Text().Contains("Previous").
			Find("ul.pagination > li:nth-child(2).page-item.active").Text().Contains("1").And().
			Test("pagination", func(driver selenium.WebDriver) error {
				return testPagination(driver, []string{"Previous", "1", "2", "3", "4", "Next"})
			}).End()

		if err != nil {
			t.Fatal(err)
		}

		addPage()
		err = seq.Get(uri.String()).
			Find(".admin-logs > table.table > tbody > tr").Count(30).
			Find("ul.pagination > li.page-item.disabled").Text().Contains("Previous").
			Find("ul.pagination > li:nth-child(2).page-item.active").Text().Contains("1").And().
			Test("pagination", func(driver selenium.WebDriver) error {
				return testPagination(driver, []string{"Previous", "1", "2", "3", "4", "5", "Next"})
			}).End()

		if err != nil {
			t.Fatal(err)
		}

		addPage()
		err = seq.Get(uri.String()).
			Find(".admin-logs > table.table > tbody > tr").Count(30).
			Find("ul.pagination > li.page-item.disabled").Text().Contains("Previous").
			Find("ul.pagination > li:nth-child(2).page-item.active").Text().Contains("1").And().
			Test("pagination", func(driver selenium.WebDriver) error {
				return testPagination(driver,
					[]string{"Previous", "1", "2", "3", "4", "5", "6", "Next"})
			}).End()

		if err != nil {
			t.Fatal(err)
		}

		addPage()
		err = seq.Get(uri.String()).
			Find(".admin-logs > table.table > tbody > tr").Count(30).
			Find("ul.pagination > li.page-item.disabled").Text().Contains("Previous").
			Find("ul.pagination > li:nth-child(2).page-item.active").Text().Contains("1").And().
			Test("pagination", func(driver selenium.WebDriver) error {
				return testPagination(driver,
					[]string{"Previous", "1", "2", "3", "4", "5", "6", "7", "Next"})
			}).End()

		if err != nil {
			t.Fatal(err)
		}

		addPage()
		err = seq.Get(uri.String()).
			Find(".admin-logs > table.table > tbody > tr").Count(30).
			Find("ul.pagination > li.page-item.disabled").Text().Contains("Previous").
			Find("ul.pagination > li:nth-child(2).page-item.active").Text().Contains("1").And().
			Test("pagination", func(driver selenium.WebDriver) error {
				return testPagination(driver,
					[]string{"Previous", "1", "2", "3", "4", "5", "6", "...", "8", "Next"})
			}).End()

		if err != nil {
			t.Fatal(err)
		}

		addPage()
		err = seq.Get(uri.String()).
			Find(".admin-logs > table.table > tbody > tr").Count(30).
			Find("ul.pagination > li.page-item.disabled").Text().Contains("Previous").
			Find("ul.pagination > li:nth-child(2).page-item.active").Text().Contains("1").And().
			Test("pagination", func(driver selenium.WebDriver) error {
				return testPagination(driver,
					[]string{"Previous", "1", "2", "3", "4", "5", "6", "...", "9", "Next"})
			}).End()

		if err != nil {
			t.Fatal(err)
		}

		addPage()
		err = seq.Get(uri.String()).
			Find(".admin-logs > table.table > tbody > tr").Count(30).
			Find("ul.pagination > li.page-item.disabled").Text().Contains("Previous").
			Find("ul.pagination > li:nth-child(2).page-item.active").Text().Contains("1").And().
			Test("pagination", func(driver selenium.WebDriver) error {
				return testPagination(driver,
					[]string{"Previous", "1", "2", "3", "4", "5", "6", "...", "10", "Next"})
			}).
			Find("ul.pagination > li:nth-child(3).page-item").Text().Contains("2").Click().And().
			Test("pagination", func(driver selenium.WebDriver) error {
				return testPagination(driver,
					[]string{"Previous", "1", "2", "3", "4", "5", "6", "...", "10", "Next"})
			}).
			Find("ul.pagination > li.page-item.active").Text().Contains("2").
			Find("ul.pagination > li:nth-child(6).page-item").Text().Contains("5").Click().And().
			Test("pagination", func(driver selenium.WebDriver) error {
				return testPagination(driver,
					[]string{"Previous", "1", "...", "3", "4", "5", "6", "7", "...", "10", "Next"})
			}).
			Find("ul.pagination > li.page-item.active").Text().Contains("5").
			Find("ul.pagination > li:nth-child(1).page-item").Text().Contains("Previous").Click().And().
			Test("pagination", func(driver selenium.WebDriver) error {
				return testPagination(driver,
					[]string{"Previous", "1", "2", "3", "4", "5", "6", "...", "10", "Next"})
			}).
			Find("ul.pagination > li.page-item.active").Text().Contains("4").
			Find("ul.pagination > li:nth-child(10).page-item").Text().Contains("Next").Click().And().
			Test("pagination", func(driver selenium.WebDriver) error {
				return testPagination(driver,
					[]string{"Previous", "1", "...", "3", "4", "5", "6", "7", "...", "10", "Next"})
			}).
			Find("ul.pagination > li.page-item.active").Text().Contains("5").
			Find("ul.pagination > li:nth-child(10).page-item").Text().Contains("10").Click().And().
			Test("pagination", func(driver selenium.WebDriver) error {
				return testPagination(driver,
					[]string{"Previous", "1", "...", "5", "6", "7", "8", "9", "10", "Next"})
			}).
			Find("ul.pagination > li.page-item.active").Text().Contains("10").
			Find("ul.pagination > li.page-item.disabled").Text().Contains("Next").
			Find(".admin-logs > table.table > tbody > tr").Count(2).
			Find("ul.pagination > li:nth-child(1).page-item").Text().Contains("Previous").Click().And().
			Test("pagination", func(driver selenium.WebDriver) error {
				return testPagination(driver,
					[]string{"Previous", "1", "...", "5", "6", "7", "8", "9", "10", "Next"})
			}).
			Find("ul.pagination > li.page-item.active").Text().Contains("9").
			Find("ul.pagination > li:nth-child(5).page-item").Text().Contains("6").Click().And().
			Test("pagination", func(driver selenium.WebDriver) error {
				return testPagination(driver,
					[]string{"Previous", "1", "...", "4", "5", "6", "7", "8", "...", "10", "Next"})
			}).
			Find("ul.pagination > li.page-item.active").Text().Contains("6").
			End()

		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Settings", func(t *testing.T) {
		uri.Path = "/admin/settings"

		err = seq.Get(uri.String()).
			Find("#settings").Count(1).
			Find(".menu").Count(1).
			Find("ul.menu > li.menu-item > a.active").Count(1).Text().Contains("Security").
			Find("h3").Text().Contains("Security").
			Find("#PasswordMinLength").Clear().SendKeys("-30").
			Find("#PasswordMinLength + button.btn-primary").Click().
			Find(".form-group.has-error > .form-input-hint").Count(1).
			Find("#PasswordMinLength").Clear().SendKeys("12").
			Find("#PasswordMinLength + button.btn-primary").Click().
			Find(".form-group.has-error > .form-input-hint").Count(0).
			Find("#PasswordRequireNumber").Click().
			Find(".form-group.has-error > .form-input-hint").Count(0).
			Find("ul.menu > li:nth-child(4).menu-item > a").Text().Contains("Documents").Click().
			Find("ul.menu > li.menu-item > a.active").Count(1).Text().Contains("Documents").
			Find("h3").Text().Contains("Documents").
			Find("#AllowPublicDocuments").Click().
			Find(".form-group.has-error > .form-input-hint").Count(0).
			Find("ul.menu > li:nth-child(5).menu-item > a").Text().Contains("Web Server").Click().
			Find("ul.menu > li.menu-item > a.active").Count(1).Text().Contains("Web Server").
			Find("h3").Text().Contains("Web Server").
			Find("#RateLimit").Clear().SendKeys("10000").
			Find("#RateLimit + button.btn-primary").Click().
			Find(".form-group.has-error > .form-input-hint").Count(0).
			Find("ul.menu > li:nth-child(6).menu-item > a").Text().Contains("Misc").Click().
			Find("ul.menu > li.menu-item > a.active").Count(1).Text().Contains("Misc").
			Find("h3").Text().Contains("Misc").
			Find("#NonAdminIssueSubmission").Click().
			Find(".form-group.has-error > .form-input-hint").Count(0).
			Find("#settingSearch").Clear().SendKeys("rate").
			Find("h3").Text().Contains("Searching").
			Find("ul.menu > li.menu-item > a.active").Count(2).Any().Text().Contains("Security").
			Find("ul.menu > li.menu-item > a.active").Count(2).Any().Text().Contains("Web Server").
			Find(".columns > .column > form.setting-group").Count(2).
			Find("#settingSearch").Clear().SendKeys("shouldn't match on anything").
			Find(".columns > .column > form.setting-group").Count(0).
			End()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Users", func(t *testing.T) {
		uri.Path = "/admin/"

		err = seq.Get(uri.String()).
			Find(".tab > .tab-item > a[href='/admin/users?active']").Click().
			Find(".admin-users").Count(1).
			Find(".btn-group > .active").Text().Equals("Active Only").
			//TODO: Finish tests
			End()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Registration", func(t *testing.T) {
		uri.Path = "/admin/registration"

		// new registration
		err = seq.Get(uri.String()).
			Find(".registration").Count(1).
			Find(".registration > a[href='/admin/newregistration']").Click().
			Find("#newRegistration").Count(1).
			Find("form  button[type='submit']").Click().
			Find(".form-group.has-error  .form-input-hint").Text().Contains("A description is required").
			Find("#tokenDescription").Clear().SendKeys("test with limit").
			Find("#tokenLimit input[type='number']").Disabled().
			Find("#tokenLimit .form-switch").Click().
			Find("#tokenLimit input[type='number']").Enabled().
			Find("form  button[type='submit']").Click().
			Find(".form-group.has-error  .form-input-hint").Text().Contains("Limit must be greater than 0").
			Find("#tokenLimit input[type='number']").Clear().SendKeys("5").
			Find("#tokenExpiration input[type='date']").Disabled().
			Find("#tokenExpiration .form-switch").Click().
			Find("#tokenExpiration input[type='date']").Enabled().
			Find("form  button[type='submit']").Click().
			Find(".form-group.has-error  .form-input-hint").Text().Contains("Please specify a date").
			Find("#tokenExpiration input[type='date']").SendKeys(dateInput(time.Now().AddDate(-1, 0, 0))).
			Find("form  button[type='submit']").Click().
			Find(".form-group.has-error  .form-input-hint").Text().Contains("Date must be after today").
			Find("#tokenExpiration input[type='date']").SendKeys(dateInput(time.Now().AddDate(1, 0, 0))).
			Find("form  button[type='submit']").Click().
			Find(".form-group.has-error  .form-input-hint").Count(0).
			End()
		if err != nil {
			t.Fatal(err)
		}

		// group testing
		gSearch := "#groupSearch > div > input"
		gResult := "#groupSearch > ul.menu > li"
		uri.Path = "/admin/newregistration"
		err = seq.Get(uri.String()).
			Find("#tokenDescription").Clear().SendKeys("group testing").
			Find(gSearch).Clear().SendKeys("test group name").Wait(500 * time.Millisecond).
			Find(gResult).Count(2).Any().Text().Contains("Create group test group name").
			Find(gResult + " > a").Click().
			Find("#newRegistration  .chips > .chip").Count(1).Text().Contains("test group name").
			Find(gSearch).Clear().SendKeys("test").Wait(500 * time.Millisecond).
			Find(gResult + ".menu-item").Count(2).Any().Text().Contains("test group name").
			Find(gResult + ":nth-last-child(1)").Text().Contains("test").Click().
			Find("#newRegistration  .chips > .chip").Count(2).All().Text().Contains("test").
			Find("#newRegistration  .chips > span:nth-child(1).chip > a.btn-clear").Click().
			Find("#newRegistration  .chips > .chip").Count(1).Text().Equals("test").
			Find(gSearch).Clear().SendKeys("new group").Wait(500 * time.Millisecond).
			Find(gResult + ":nth-last-child(1)").Click().
			Find("#newRegistration  .chips > .chip").Count(2).
			And().Get(uri.String()).
			Find("#tokenDescription").Clear().SendKeys("group testing").
			Find(gSearch).Clear().SendKeys("test").Wait(500 * time.Millisecond).
			Find(gResult + ".menu-item").Count(2).All().Text().Contains("test").
			Find(gSearch).SendKeys("new group").Wait(500 * time.Millisecond).
			Find(gResult + ".menu-item").Count(1).Text().Contains("new group").
			And().Get(uri.String()).
			Find("#tokenDescription").Clear().SendKeys("group testing").
			Find(gSearch).Clear().SendKeys("group").Wait(500 * time.Millisecond).
			Find(gResult + ".menu-item").Count(3).All().Text().Contains("group").
			Any().
			Text().Contains("new group").
			Text().Contains("test group name").
			Text().Contains("Create group group").
			Find(gResult + ":nth-child(1).menu-item").Click().
			Find(gSearch).Clear().SendKeys("group").Wait(500 * time.Millisecond).
			Find(gResult + ":nth-child(2).menu-item").Click().
			Find(gSearch).Clear().SendKeys("group").Wait(500 * time.Millisecond).
			Find(gResult + ":nth-child(3).menu-item").Click().
			Find("#newRegistration  .chips > .chip").Count(3).
			Find("form  button[type='submit']").Click().
			Find(".form-group.has-error  .form-input-hint").Count(0).
			End()
		if err != nil {
			t.Fatal(err)
		}

		uri.Path = "/admin/registration"
		// registration list
		err = seq.URL().Path("/admin/registration").
			Find(".registration > .table > tbody > tr").Count(2).
			Find(".registration > a[href='/admin/newregistration']").Click().
			Find("#tokenDescription").Clear().SendKeys("Valid Token Test").
			Find("form  button[type='submit']").Click().
			Find(".registration > .table > tbody > tr").Count(3).
			Filter(func(e *sequence.Elements) error {
				return e.Text().Contains("Valid Token Test").End()
			}).Count(1).FindChildren("td:nth-child(1) > a").Click().
			Find("button.btn.btn-error").Text().Contains("Remove").Click().
			And().Get(uri.String()).
			Find(".registration > .table > tbody > tr").Count(2).
			Find(".btn-group > a.btn").Count(2).
			Find(".btn-group > a.btn.active").Text().Equals("Active").
			Find(".btn-group > a.btn:not(.active)").Text().Equals("All").Click().
			Find(".registration > .table > tbody > tr").Count(3).
			Find(".registration > .table > tbody > tr.secondary").Text().Contains("Valid Token Test").
			End()

		if err != nil {
			t.Fatal(err)
		}

		// view single registration
		err = seq.URL().Path("/admin/registration").
			Find(".registration > .table > tbody > tr > td > a").
			Filter(func(e *sequence.Elements) error {
				return e.Text().Contains("group testing").End()
			}).Click().
			Find("#singleRegistration h4").Text().Contains("group testing").
			Find("#singleRegistration .tile-title").Text().Contains("Created by " + username).
			Find("#singleRegistration .chip").Count(3).Any().
			Text().Contains("test group name").
			Text().Contains("new group").
			Text().Contains("test").
			Text().Contains("group").
			Find("#singleRegistration .input-group > button.btn-primary").Click().
			Find("#singleRegistration button.btn.btn-error").Text().Contains("Remove").Click().
			Find("#singleRegistration .input-group > input.form-input").Count(0).
			End()
		if err != nil {
			t.Fatal(err)
		}

		tokenUrl := ""

		testUsername := "registered"
		testPassword := "registeredPasswordThatHasANumber4"

		// test registration link
		err = seq.Get(uri.String()).
			Find(".registration > .table > tbody > tr > td:nth-child(1) > a").Count(1).Click().
			Find("#singleRegistration .tile-title").Text().Contains("Created by "+username).
			Find("#singleRegistration .tile-content .tile-subtitle.text-gray").
			Any().Text().Contains("5 Registrations Left").
			Find("#singleRegistration .input-group > input").
			Test("get token url", func(e selenium.WebElement) error {
				u, err := e.GetAttribute("value")
				if err != nil {
					return err
				}
				tokenUrl = u
				return nil
			}).
			And().Get(tokenUrl).
			Find("#signup #inputUsername").SendKeys(testUsername).
			Find("#signup #inputPassword").SendKeys(testPassword).
			Find("#signup #inputPassword2").SendKeys(testPassword).
			Find("#submit").Click().
			Find(".form-input-hint").Count(0).
			And().Get(uri.String()).
			Find("#inputUsername").SendKeys(username).
			Find("#inputPassword").SendKeys(password).
			Find(".btn.btn-primary.btn-block").Click().
			Find(".registration > .table > tbody > tr > td:nth-child(1) > a").Count(1).Click().
			Find("#singleRegistration .tile-content .tile-subtitle.text-gray").
			Any().Text().Contains("4 Registrations Left").
			Find(".tile-row > .tile").Count(1).
			FindChildren(".tile-content > .tile-title").Text().Contains(testUsername).
			End()

		if err != nil {
			t.Fatal(err)
		}

	})
}
