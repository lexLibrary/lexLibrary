// Copyright (c) 2017-2018 Townsourced Inc.

package web_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
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

	uri.Path = "/admin/logs"

	addPage()
	err = seq.Get(uri.String()).
		Find(".logs > table.table > tbody > tr").Count(30).
		Find("ul.pagination > li.page-item.disabled").Text().Contains("Previous").
		Find("ul.pagination > li:nth-child(2).page-item.active").Text().Contains("1").
		Find("ul.pagination > li:nth-child(3).page-item").Text().Contains("2").
		Find("ul.pagination > li:nth-child(4).page-item").Text().Contains("Next").
		End()
	if err != nil {
		t.Fatal(err)
	}

	addPage()
	err = seq.Get(uri.String()).
		Find(".logs > table.table > tbody > tr").Count(30).
		Find("ul.pagination > li:nth-child(2).page-item.active").Text().Contains("1").
		Find("ul.pagination > li:nth-child(3).page-item").Text().Contains("2").
		Find("ul.pagination > li:nth-child(4).page-item").Any().Text().Contains("3").
		Find("ul.pagination > li:nth-child(5).page-item").Text().Contains("Next").
		End()
	if err != nil {
		t.Fatal(err)
	}

	// Settings
	// Registration
}
