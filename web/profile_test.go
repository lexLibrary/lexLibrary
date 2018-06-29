// Copyright (c) 2017-2018 Townsourced Inc.
package web_test

import (
	"strings"
	"testing"
)

func TestProfile(t *testing.T) {
	uri := *llURL
	uri.Path = "profile"

	username := "testUser"
	password := "testpasswordThatisLongEnough"

	setupUserAndLogin(t, username, password, false)

	err := newSequence().Get(uri.String()).
		Find(".profile-edit").Count(1).
		Find("figure.avatar.avatar-full").Count(1).Attribute("data-initial").Equals("TE").
		Find("#displayName").Text().Equals(strings.ToLower(username)).
		Find("ul.tab.tab-block > li.tab-item.active").Count(1).Text().Contains("Documents").
		Find("a[href='/profile/readLater']").Count(1).Click().
		Find("ul.tab.tab-block > li.tab-item.active").Count(1).Text().Contains("Read Later").
		Find("a[href='/profile/comments']").Count(1).Click().
		Find("ul.tab.tab-block > li.tab-item.active").Count(1).Text().Contains("Comments").
		Find("a[href='/profile/history']").Count(1).Click().
		Find("ul.tab.tab-block > li.tab-item.active").Count(1).Text().Contains("History").
		End()
	if err != nil {
		t.Fatal(err)
	}

	// profile edit
	err = newSequence().
		Find("a[href='/profile/edit'].profile-edit").Count(1).Click().
		And().Title().Contains("Edit Profile").
		Find("figure.avatar.avatar-full").Attribute("data-initial").Equals("TE").
		Find("#inputName").SendKeys("Test A New Name").
		Find("#changeName > button.btn").Click().
		Find("figure.avatar.avatar-full").Attribute("data-initial").Equals("TN").
		End()
	if err != nil {
		t.Fatal(err)
	}

	// profile change password
	err = newSequence().
		Find("a[href='/profile/edit/account']").Click().
		Find("#changePassword > button.btn").Click().
		Find("#changePassword > .has-error > .form-input-hint").
		Text().Contains("You must enter your old password").
		Find("#inputPasswordOld").SendKeys(password).
		Find("#changePassword > button.btn").Click().
		Find("#changePassword > .has-error > .form-input-hint").
		Text().Contains("You must enter a new password").
		Find("#inputPasswordNew").SendKeys(password + "new").
		Find("#changePassword > button.btn").Click().
		Find("#changePassword > .has-error > .form-input-hint").
		Text().Contains("Passwords do not match").
		Find("#inputPasswordConfirm").SendKeys(password + "somethingelse").
		Find("#changePassword > button.btn").Click().
		Find("#changePassword > .has-error > .form-input-hint").
		Text().Contains("Passwords do not match").
		Find("#inputPasswordConfirm").SendKeys(password + "new").
		Find("#changePassword > button.btn").Click().
		End()
	if err != nil {
		t.Fatal(err)
	}

	// change username
	err = newSequence().
		Find("#inputUsername").Clear().SendKeys("ts").
		Find("#changeUsername > button.btn").Click().
		Find("#changeUsername > .has-error > .form-input-hint").
		Text().Contains("username must be greater than").
		Find("#inputUsername").SendKeys("newusername").
		Find("#changeUsername > button.btn").Click().
		Find("#changeUsername > .has-error > .form-input-hint").Count(0).
		End()
	if err != nil {
		t.Fatal(err)
	}

}
