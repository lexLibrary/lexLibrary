// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
)

func TestGroup(t *testing.T) {

	var admin *app.Admin
	var user *app.User
	reset := func(t *testing.T) {
		t.Helper()
		admin = resetAdmin(t, "admin", "newuserpassword")

		err := admin.SetSetting("AllowPublicSignups", true)
		if err != nil {
			t.Fatalf("Error allowing public signups for testing: %s", err)
		}
		user, err = app.UserNew("newuser", "newuserpassword")
		if err != nil {
			t.Fatalf("Error adding user: %s", err)
		}

	}

	t.Run("New", func(t *testing.T) {
		reset(t)
		_, err := user.NewGroup(fmt.Sprintf("Z%128s", "test group"))
		if err == nil {
			t.Fatal("Adding a new group didn't limit the group name size")
		}
		_, err = user.NewGroup("")
		if !app.IsFail(err) {
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

		_, err = user.NewGroup(strings.ToUpper(g.Name))
		if err == nil {
			t.Fatalf("Adding a new group with an existing group's case insensitive name didn't fail")
		}
	})
	t.Run("Admin", func(t *testing.T) {
		reset(t)
		g, err := admin.User().NewGroup("New group")
		if err != nil {
			t.Fatalf("Error creating group: %s", err)
		}

		_, err = g.Admin(nil)
		if !app.IsFail(err) {
			t.Fatalf("Getting admin with nil user did not fail")
		}

		_, err = g.Admin(user)
		if !app.IsFail(err) {
			t.Fatalf("Getting admin from a non member user did not fail")
		}

		ga, err := g.Admin(admin.User())
		if err != nil {
			t.Fatalf("Error getting group admin: %s", err)
		}

		err = ga.AddMember(user.ID)
		if err != nil {
			t.Fatalf("Error adding member to group: %s", err)
		}

		_, err = g.Admin(user)
		if !app.IsFail(err) {
			t.Fatalf("Getting admin from a non admin member did not fail")
		}

		err = ga.AddMember(admin.User().ID)
		if err != nil {
			t.Fatalf("Error removing admin from group: %s", err)
		}

		_, err = g.Admin(admin.User())
		if err != nil {
			t.Fatalf("Site admin did not have implicit admin permissions on groups: %s", err)
		}
	})

	t.Run("Set Name", func(t *testing.T) {
		reset(t)
		g, err := user.NewGroup("New group")
		if err != nil {
			t.Fatalf("Error creating group: %s", err)
		}

		ga, err := g.Admin(user)
		if err != nil {
			t.Fatalf("Error getting group admin: %s", err)
		}

		err = ga.SetName(fmt.Sprintf("Z%128s name", "test group"), g.Version)
		if !app.IsFail(err) {
			t.Fatal("Setting group name didn't limit the group name size")
		}
		name := "New Group Name"

		err = ga.SetName(name, 4)
		if !app.IsFail(err) {
			t.Fatalf("Setting group name with invalid version did not fail")
		}

		err = ga.SetName(name, g.Version)
		if err != nil {
			t.Fatalf("Error setting group name: %s", err)
		}

		newName := ""
		err = data.NewQuery(`select name from groups where id = {{arg "id"}}`).QueryRow(data.Arg("id", g.ID)).
			Scan(&newName)
		if err != nil {
			t.Fatalf("Error getting new group name: %s", err)
		}

		if newName != name {
			t.Fatalf("Group name is incorrect.  Expected %s, got %s", name, newName)
		}
	})

	t.Run("Set Member", func(t *testing.T) {
		reset(t)
		g, err := user.NewGroup("New group")
		if err != nil {
			t.Fatalf("Error creating group: %s", err)
		}

		ga, err := g.Admin(user)
		if err != nil {
			t.Fatalf("Error getting group admin: %s", err)
		}

		other, err := app.UserNew("otheruser", "newuserpassword")
		if err != nil {
			t.Fatalf("Error adding user: %s", err)
		}

		id := data.ID{}

		err = ga.AddMember(id)
		if !app.IsFail(err) {
			t.Fatalf("No failure when adding an invalid group member id")
		}

		err = ga.AddMember(data.NewID())
		if !app.IsFail(err) {
			t.Fatalf("No failure when adding an invalid group member")
		}

		err = ga.AddMember(other.ID)
		if err != nil {
			t.Fatalf("Error adding new group member: %s", err)
		}

		isAdmin := false
		getMember := data.NewQuery(`select admin from group_users
			where group_id = {{arg "group_id"}} and user_id = {{arg "user_id"}}`)

		err = getMember.QueryRow(data.Arg("user_id", other.ID), data.Arg("group_id", g.ID)).Scan(&isAdmin)
		if err != nil {
			t.Fatalf("Error getting group member: %s", err)
		}
		if err == sql.ErrNoRows {
			t.Fatalf("No member found")
		}

		err = ga.AddAdmin(other.ID)
		if err != nil {
			t.Fatalf("Error updating group member: %s", err)
		}

		err = getMember.QueryRow(data.Arg("user_id", other.ID), data.Arg("group_id", g.ID)).Scan(&isAdmin)
		if err != nil {
			t.Fatalf("Error getting group member: %s", err)
		}
		if err == sql.ErrNoRows {
			t.Fatalf("No member found")
		}

		if !isAdmin {
			t.Fatalf("Invalid group admin value. Expected %t, got %t", true, isAdmin)
		}

		err = ga.AddAdmin(other.ID)
		if err != nil {
			t.Fatalf("Error updating group member: %s", err)
		}

		err = ga.AddAdmin(other.ID)
		if err != nil {
			t.Fatalf("Error updating group member who is already a member: %s", err)
		}
	})

	t.Run("GroupSearch", func(t *testing.T) {
		reset(t)
		groups := []string{
			"Group One",
			"Group Two",
			"Another Group",
			"a Matching group name",
			"Something else",
		}

		tests := []struct {
			search  string
			results []string
		}{
			{"Group", []string{"Group One", "Group Two", "Another Group", "a Matching group name"}},
			{"group", []string{"Group One", "Group Two", "Another Group", "a Matching group name"}},
			{"a", []string{"Another Group", "a Matching group name"}},
			{"A", []string{"Another Group", "a Matching group name"}},
			{"Som", []string{"Something else"}},
		}

		for i := range groups {
			_, err := user.NewGroup(groups[i])
			if err != nil {
				t.Fatalf("Error creating group %s ERROR: %s", groups[i], err)
			}
		}

		for i, test := range tests {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				groups, err := user.GroupSearch(test.search)
				if err != nil {
					t.Fatalf("Error running group search: %s", err)
				}

				if len(groups) != len(test.results) {
					t.Fatalf("Incorrect group search result length. Expected %d, got %d",
						len(test.results), len(groups))
				}

				for _, result := range test.results {
					found := false
					for _, group := range groups {
						if group.Name == result {
							found = true
							break
						}
					}
					if !found {
						t.Fatalf("Result '%s' was not found in group search results for "+
							"test '%s'", result, test.search)
					}
				}
			})
		}

	})
}
