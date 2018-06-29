// Copyright (c) 2017-2018 Townsourced Inc.

package web_test

import (
	"path"
	"strings"
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
)

func TestUser(t *testing.T) {
	username := "testuser"
	password := "testpasswordThatisLongEnough"

	setupUserAndLogin(t, username, password, true)

	adminUser, err := app.Login(username, password)
	if err != nil {
		t.Fatal(err)
	}
	_, err = adminUser.Admin()
	if err != nil {
		t.Fatal(err)
	}

	other, err := app.UserNew("otheruser", password)
	if err != nil {
		t.Fatal(err)
	}

	seq := newSequence()

	uri := *llURL
	uri.Path = path.Join("user", other.Username)

	err = seq.Get(uri.String()).
		Find(".profile-edit").Count(0).
		Find("figure.avatar.avatar-full").Count(1).Attribute("data-initial").Equals("OT").
		Find("#displayName").Text().Equals(strings.ToLower(other.Username)).
		Find("ul.tab.tab-block > li.tab-item.active").Count(1).Text().Contains("Documents").
		Find("ul.tab.tab-block > li:nth-child(2).tab-item").Count(1).Text().Contains("Comments").Click().
		Find("ul.tab.tab-block > li.tab-item.active").Count(1).Text().Contains("Comments").
		End()
	if err != nil {
		t.Fatal(err)
	}

	uri.Path = path.Join("user", adminUser.Username)

	err = seq.Get(uri.String()).
		Find(".profile-edit").Count(1).
		Find("figure.avatar.avatar-full").Count(1).Attribute("data-initial").Equals("TE").
		Find("#displayName").Text().Contains(strings.ToLower(adminUser.Username)).
		Find("ul.tab.tab-block > li.tab-item.active").Count(1).Text().Contains("Documents").
		Find("ul.tab.tab-block > li:nth-child(2).tab-item").Count(1).Text().Contains("Read Later").Click().
		Find("ul.tab.tab-block > li.tab-item.active").Count(1).Text().Contains("Read Later").
		Find("ul.tab.tab-block > li:nth-child(3).tab-item").Count(1).Text().Contains("Comments").Click().
		Find("ul.tab.tab-block > li.tab-item.active").Count(1).Text().Contains("Comments").
		Find("ul.tab.tab-block > li:nth-child(4).tab-item").Count(1).Text().Contains("History").Click().
		Find("ul.tab.tab-block > li.tab-item.active").Count(1).Text().Contains("History").
		Find("section.navbar-section > a.btn.btn-link").Text().Contains("Logout").
		End()
	if err != nil {
		t.Fatal(err)
	}

}
