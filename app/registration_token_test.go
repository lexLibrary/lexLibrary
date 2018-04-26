// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"testing"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
)

func TestRegistrationToken(t *testing.T) {
	var admin *app.Admin
	reset := func(t *testing.T) {
		t.Helper()

		admin = resetAdmin(t, "admin", "newuserpassword").AsAdmin()
		err := admin.SetSetting("AllowPublicSignups", true)
		if err != nil {
			t.Fatalf("Error allowing public signups for testing: %s", err)
		}
	}

	t.Run("New", func(t *testing.T) {
		reset(t)

		_, err := admin.NewRegistrationToken(0, time.Now().AddDate(0, 0, -10), nil)
		if !app.IsFail(err) {
			t.Fatalf("Generating token with old expire date didn't fail")
		}

		_, err = admin.NewRegistrationToken(0, time.Time{}, []data.ID{data.NewID()})
		if !app.IsFail(err) {
			t.Fatalf("Generating a token with an invalid groupID did not fail")
		}

		group, err := admin.User.NewGroup("New Test Group")
		if err != nil {
			t.Fatalf("Error creating group for testing")
		}

		group2, err := admin.User.NewGroup("New Test Group2")
		if err != nil {
			t.Fatalf("Error creating group for testing")
		}

		_, err = admin.NewRegistrationToken(0, time.Time{}, []data.ID{group.ID, group2.ID, data.NewID()})
		if !app.IsFail(err) {
			t.Fatalf("Generating a token with at least one invalid groupID did not fail: %s", err)
		}

		token, err := admin.NewRegistrationToken(0, time.Time{}, nil)
		if err != nil {
			t.Fatalf("Generating registration token failed: %s", err)
		}
		if token.Token == "" {
			t.Fatalf("Invalid token: %s", token.Token)
		}

		if token.Expires.Valid || !token.Expires.Time.IsZero() {
			t.Fatalf("Invalid null expiration value. Expected %v, got %v", time.Time{}, token.Expires)
		}

		if token.Limit != -1 {
			t.Fatalf("Invalid null limit value. Expected %d got %d", -1, token.Limit)
		}

		if len(token.Groups) != 0 {
			t.Fatalf("Invalid empty group list. Expected len %d, got %d", 0, len(token.Groups))
		}

	})

	t.Run("Register User From Token", func(t *testing.T) {
		reset(t)

		_, err := app.RegisterUserFromToken("newuser", "newuserPassword", "GarbageToken")
		if !app.IsFail(err) {
			t.Fatal("Registering user with invalid token did not fail")
		}

		token, err := admin.NewRegistrationToken(0, time.Time{}, nil)
		if err != nil {
			t.Fatalf("Generating registration token failed: %s", err)
		}
		_, err = app.RegisterUserFromToken(admin.User.Username, "newuserPassword", token.Token)
		if !app.IsFail(err) {
			t.Fatal("Registering user with an existing username didn't fail")
		}

		newUsername := "newuser"

		u, err := app.RegisterUserFromToken(newUsername, "newuserPassword", token.Token)
		if err != nil {
			t.Fatalf("Error registering new user from token: %s", err)
		}

		if u == nil {
			t.Fatal("Registered user is nil")
		}

		if u.Username != newUsername {
			t.Fatalf("Registered username didn't match submitted username. Expected %s got %s", newUsername,
				u.Username)
		}

	})

	t.Run("Limit", func(t *testing.T) {
		reset(t)

		token, err := admin.NewRegistrationToken(3, time.Time{}, nil)
		if err != nil {
			t.Fatalf("Generating registration token failed: %s", err)
		}

		_, err = app.RegisterUserFromToken("user1", "newuserPassword", token.Token)
		if err != nil {
			t.Fatalf("Error registering new user from token: %s", err)
		}

		_, err = app.RegisterUserFromToken("user2", "newuserPassword", token.Token)
		if err != nil {
			t.Fatalf("Error registering new user from token: %s", err)
		}

		_, err = app.RegisterUserFromToken("user3", "newuserPassword", token.Token)
		if err != nil {
			t.Fatalf("Error registering new user from token: %s", err)
		}

		_, err = app.RegisterUserFromToken("user4", "newuserPassword", token.Token)
		if !app.IsFail(err) {
			t.Fatalf("Registering more users than limit did not fail: %s", err)
		}
	})

	t.Run("Expiration", func(t *testing.T) {
		reset(t)
		token, err := admin.NewRegistrationToken(0, time.Now().Add(2*time.Second), nil)
		if err != nil {
			t.Fatalf("Generating registration token failed: %s", err)
		}

		_, err = app.RegisterUserFromToken("user", "newuserPassword", token.Token)
		if err != nil {
			t.Fatalf("Error registering new user from token: %s", err)
		}

		time.Sleep(4 * time.Second)

		_, err = app.RegisterUserFromToken("userfail", "newuserPassword", token.Token)
		if !app.IsFail(err) {
			t.Fatalf("Registering user after token expired did not fail: %s", err)
		}
	})

	t.Run("Groups", func(t *testing.T) {
		reset(t)
		g, err := admin.User.NewGroup("Test Group 1")
		if err != nil {
			t.Fatalf("Error adding a new group: %s", err)
		}

		g2, err := admin.User.NewGroup("Test Group 2")
		if err != nil {
			t.Fatalf("Error adding a new group: %s", err)
		}

		token, err := admin.NewRegistrationToken(0, time.Time{}, []data.ID{g.ID, g2.ID})
		if err != nil {
			t.Fatalf("Generating registration token failed: %s", err)
		}

		u, err := app.RegisterUserFromToken("user", "newuserPassword", token.Token)
		if err != nil {
			t.Fatalf("Error registering new user from token: %s", err)
		}

		rows, err := data.NewQuery(`select group_id from user_to_groups where user_id = {{arg "id"}}`).
			Query(data.Arg("id", u.ID))
		if err != nil {
			t.Fatalf("Error looking up user groups: %s", err)
		}
		defer rows.Close()

		ids := make([]data.ID, 0, 2)

		for rows.Next() {
			var id data.ID
			err = rows.Scan(&id)
			if err != nil {
				t.Fatalf("Error scanning group ID: %s", err)
			}
			ids = append(ids, id)
		}

		if len(ids) != 2 {
			t.Fatalf("User is not a member the right number of groups.  Expected %d, got %d", 2, len(ids))
		}

		for i := range ids {
			if ids[i] != g.ID && ids[i] != g2.ID {
				t.Fatalf("Group ID not found in user's membership. ID: %s ", ids[i])
			}
		}

	})
}
