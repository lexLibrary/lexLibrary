// Copyright (c) 2017-2018 Townsourced Inc.

package web_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/tebeka/selenium"
	"github.com/timshannon/sequence"
)

func TestAdmin(t *testing.T) {
	uri := *llURL
	uri.Path = "admin"

	username := "testUser"
	password := "testpasswordThatisLongEnough"

	err := createUserAndLogin(username, password, true)
	if err != nil {
		t.Fatalf("Error setting up user for testing: %s", err)
	}

	seq := newSequence().Get(uri.String())

	// overview
	err = seq.
		Find(".tab > .tab-item.active").Count(1).Text().Contains("Overview").
		Find(".overview.table").Count(6).
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

	// Logs
	errMessage := "New Error message"
	errID := app.LogError(fmt.Errorf(errMessage))

	err = seq.
		Find(".tab > .tab-item > a[href='/admin/logs']").Click().
		Find(".tab > .tab-item.active").Count(1).Text().Contains("Logs").
		Find(".logs > table.table > tbody > tr").Text().Contains(errMessage).
		Find(".logs > table.table > tbody > tr > td > a.float-right").Text().Contains("View").Click().
		Find("h4").Text().Contains("Log Entry from").
		Find("h4 > small").Text().Contains(errID.String()).
		Find(".logs.section > p").Text().Contains(errMessage).
		End()
	if err != nil {
		t.Fatal(err)
	}

	// searching
	app.LogError(fmt.Errorf("other error"))
	err = seq.Back().Refresh().
		Find(".logs > table.table > tbody > tr").Count(2).
		Find(".input-group > input.input[type='text']").SendKeys(strings.ToUpper(errMessage)).
		Find(".input-group > button.btn.btn-primary").Click().
		Find(".logs > table.table > tbody > tr").Count(1).Text().Contains(errMessage).
		Find(".input-group > input.input[type='text']").SendKeys(errID.String()).
		Find(".input-group > button.btn.btn-primary").Click().
		Find("h4").Text().Contains("Log Entry from").
		Find("h4 > small").Text().Contains(errID.String()).
		Find(".logs.section > p").Text().Contains(errMessage).
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
		Find(".logs > table.table > tbody > tr").Count(30).
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
		Find(".logs > table.table > tbody > tr").Count(30).
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
		Find(".logs > table.table > tbody > tr").Count(30).
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
		Find(".logs > table.table > tbody > tr").Count(30).
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
		Find(".logs > table.table > tbody > tr").Count(30).
		Find("ul.pagination > li.page-item.disabled").Text().Contains("Previous").
		Find("ul.pagination > li:nth-child(2).page-item.active").Text().Contains("1").And().
		Test("pagination", func(driver selenium.WebDriver) error {
			return testPagination(driver, []string{"Previous", "1", "2", "3", "4", "5", "6", "Next"})
		}).End()

	if err != nil {
		t.Fatal(err)
	}

	addPage()
	err = seq.Get(uri.String()).
		Find(".logs > table.table > tbody > tr").Count(30).
		Find("ul.pagination > li.page-item.disabled").Text().Contains("Previous").
		Find("ul.pagination > li:nth-child(2).page-item.active").Text().Contains("1").And().
		Test("pagination", func(driver selenium.WebDriver) error {
			return testPagination(driver, []string{"Previous", "1", "2", "3", "4", "5", "6", "7", "Next"})
		}).End()

	if err != nil {
		t.Fatal(err)
	}

	addPage()
	err = seq.Get(uri.String()).
		Find(".logs > table.table > tbody > tr").Count(30).
		Find("ul.pagination > li.page-item.disabled").Text().Contains("Previous").
		Find("ul.pagination > li:nth-child(2).page-item.active").Text().Contains("1").And().
		Test("pagination", func(driver selenium.WebDriver) error {
			return testPagination(driver, []string{"Previous", "1", "2", "3", "4", "5", "6", "...", "8", "Next"})
		}).End()

	if err != nil {
		t.Fatal(err)
	}

	addPage()
	err = seq.Get(uri.String()).
		Find(".logs > table.table > tbody > tr").Count(30).
		Find("ul.pagination > li.page-item.disabled").Text().Contains("Previous").
		Find("ul.pagination > li:nth-child(2).page-item.active").Text().Contains("1").And().
		Test("pagination", func(driver selenium.WebDriver) error {
			return testPagination(driver, []string{"Previous", "1", "2", "3", "4", "5", "6", "...", "9", "Next"})
		}).End()

	if err != nil {
		t.Fatal(err)
	}

	addPage()
	err = seq.Get(uri.String()).
		Find(".logs > table.table > tbody > tr").Count(30).
		Find("ul.pagination > li.page-item.disabled").Text().Contains("Previous").
		Find("ul.pagination > li:nth-child(2).page-item.active").Text().Contains("1").And().
		Test("pagination", func(driver selenium.WebDriver) error {
			return testPagination(driver, []string{"Previous", "1", "2", "3", "4", "5", "6", "...", "10", "Next"})
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
		Find(".logs > table.table > tbody > tr").Count(2).
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

	// Settings
	uri.Path = "/admin/settings"

	addPage()
	err = seq.Get(uri.String()).
		Find("#settings").Count(1).
		Find(".menu").Count(1).
		Find("ul.menu > li.menu-item > a.active").Count(1).Text().Contains("Security").
		Find("h3").Text().Contains("Security").
		Find("#PasswordMinLength").Clear().SendKeys("-30").
		Find("#PasswordMinLength + button.btn-primary").Click().
		Find(".form-group.has-error > .form-input-hint").Count(1).
		Find("#PasswordMinLength").Clear().SendKeys("30").
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

	// Registration
}
