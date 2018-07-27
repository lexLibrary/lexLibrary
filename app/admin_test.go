// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
)

func TestAdmin(t *testing.T) {
	var admin *app.Admin

	reset := func(t *testing.T) {
		t.Helper()

		admin = resetAdmin(t, "admin", "adminpassword")
		ok(t, admin.SetSetting("AllowPublicSignups", true))

	}

	t.Run("Overview", func(t *testing.T) {
		reset(t)

		overview, err := admin.Overview()
		ok(t, err)

		assert(t, overview != nil, "Admin Overview is nil")
	})

	t.Run("InstanceUsers", func(t *testing.T) {
		reset(t)

		password := "passwordValue"

		inactive, err := app.UserNew("inactive", password)
		ok(t, err)

		err = admin.SetUserActive(inactive.Username, false)
		ok(t, err)

		loggedIn, err := app.UserNew("loggedin", password)
		ok(t, err)

		_, err = loggedIn.NewSession(time.Now().Add(1*time.Hour), "", "")
		ok(t, err)

		loggedOut, err := app.UserNew("loggedout", password)
		ok(t, err)

		s, err := loggedOut.NewSession(time.Now().Add(1*time.Hour), "", "")
		ok(t, err)
		ok(t, s.Logout())

		multipleSessions, err := app.UserNew("multiplesessions", password)
		ok(t, err)

		_, err = multipleSessions.NewSession(time.Now().Add(1*time.Hour), "", "")
		ok(t, err)
		_, err = multipleSessions.NewSession(time.Now().Add(1*time.Hour), "", "")
		ok(t, err)

		neverLoggedIn, err := app.UserNew("neverLoggedIn", password)
		ok(t, err)

		ok(t, loggedIn.SetName("John Doe", loggedIn.Version))

		ok(t, loggedOut.SetName("James Doe", loggedOut.Version))

		tests := []struct {
			activeOnly bool
			loggedIn   bool
			search     string
			offset     int
			limit      int

			total  int
			result []*app.User
		}{
			{true, false, "", 0, 100, 5, []*app.User{admin.User(), loggedIn, loggedOut, multipleSessions,
				neverLoggedIn}},
			{false, true, "", 0, 100, 2, []*app.User{loggedIn, multipleSessions}},
			{true, true, "", 0, 100, 2, []*app.User{loggedIn, multipleSessions}},
			{false, false, "", 0, 100, 6, []*app.User{admin.User(), loggedIn, loggedOut, multipleSessions,
				inactive, neverLoggedIn}},
			{false, false, "", 0, 2, 6, []*app.User{neverLoggedIn, multipleSessions}},
			{false, false, "", 2, 2, 6, []*app.User{loggedOut, loggedIn}},
			{false, false, "", 4, 2, 6, []*app.User{inactive, admin.User()}},
			{false, false, "John", 0, 100, 1, []*app.User{loggedIn}},
			{false, false, "logged", 0, 100, 3, []*app.User{loggedIn, loggedOut, neverLoggedIn}},
			{false, false, "Doe", 0, 100, 2, []*app.User{loggedIn, loggedOut}},
			{false, true, "DOE", 0, 100, 1, []*app.User{loggedIn}},
		}

		for i, test := range tests {
			t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
				users, total, err := admin.InstanceUsers(test.activeOnly, test.loggedIn, test.search,
					test.offset, test.limit)
				ok(t, err)

				equals(t, total, test.total)

				equals(t, len(users), len(test.result))

				for _, result := range test.result {
					found := false
					for _, user := range users {
						if user.Username == result.Username {
							found = true
							break
						}
					}
					assert(t, found, "User %s was not found in the result set.", result.Username)
				}
			})
		}

	})

	t.Run("InstanceUser", func(t *testing.T) {
		reset(t)

		user, err := app.UserNew("instanceUserTest", "testInstanceUserPassword")
		ok(t, err)

		iu, err := admin.InstanceUser("instanceUSERTEST")
		ok(t, err)

		equals(t, iu.Username, user.Username)
		equals(t, iu.ID, user.ID)

		assert(t, !iu.LastLogin.Valid, "Instance user's last login was valid. Expected %v got %v",
			false, iu.LastLogin.Valid)

		_, err = user.NewSession(time.Now().Add(1*time.Hour), "", "")
		ok(t, err)

		iu, err = admin.InstanceUser("instanceusertest")
		ok(t, err)
		equals(t, iu.Username, user.Username)
		equals(t, iu.ID, user.ID)
		assert(t, iu.LastLogin.Valid, "Instance user's last login was not valid. Expected %v got %v", true,
			iu.LastLogin.Valid)

	})
}
