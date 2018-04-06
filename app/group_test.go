// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"fmt"
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
)

func TestGroup(t *testing.T) {

	var user *app.User
	reset := func(t *testing.T) {
		t.Helper()

		_, err := data.NewQuery("delete from users_to_groups").Exec()
		if err != nil {
			t.Fatalf("Error emptying users_to_groups table before running tests: %s", err)
		}

		_, err = data.NewQuery("delete from groups").Exec()
		if err != nil {
			t.Fatalf("Error emptying groups table before running tests: %s", err)
		}

		user = prepUser(t, "newuser", "newuserpassword")

	}

	t.Run("New", func(t *testing.T) {
		reset(t)
		_, err := user.NewGroup(fmt.Sprintf("%70s", "test group"))
		if err == nil {
			t.Fatal("Adding a new group didn't limit the group name size")
		}
		_, err = user.NewGroup("")
		if err == nil {
			t.Fatal("Adding a new group without a name didn't fail")
		}

		g, err := user.NewGroup("New Group Name")
		if err != nil {
			t.Fatalf("Adding a new group failed: %s", err)
		}

		_, err = user.NewGroup(g.Name)
		if err == nil {
			t.Fatalf("Adding a new group with an existing group's name didn't fail")
		}

	})
}
