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

		admin = prepAdmin(t, "admin", "newuserpassword").AsAdmin()
		err := admin.SetSetting("AllowPublicSignups", true)
		if err != nil {
			t.Fatalf("Error allowing public signups for testing: %s", err)
		}
		truncateTable(t, "registration_tokens")
		truncateTable(t, "registration_token_groups")
		truncateTable(t, "user_to_groups")
		truncateTable(t, "groups")
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

		if !token.Expires.Time.IsZero() {
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

	})

	// register new with expiration
	// register new with groups
}
