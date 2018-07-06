// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
)

func TestRegistrationToken(t *testing.T) {
	var admin *app.Admin
	reset := func(t *testing.T) {
		t.Helper()

		admin = resetAdmin(t, "admin", "newuserpassword")
		err := admin.SetSetting("AllowPublicSignups", true)
		if err != nil {
			t.Fatalf("Error allowing public signups for testing: %s", err)
		}
	}

	t.Run("New", func(t *testing.T) {
		reset(t)

		_, err := admin.NewRegistrationToken("test", 0, time.Now().AddDate(0, 0, -10), nil)
		if !app.IsFail(err) {
			t.Fatalf("Generating token with old expire date didn't fail")
		}

		_, err = admin.NewRegistrationToken("test", 0, time.Time{}, []data.ID{data.NewID()})
		if !app.IsFail(err) {
			t.Fatalf("Generating a token with an invalid groupID did not fail")
		}

		_, err = admin.NewRegistrationToken("", 0, time.Time{}, []data.ID{data.NewID()})
		if !app.IsFail(err) {
			t.Fatalf("Generating a token with an invalid description did not fail")
		}

		group, err := admin.User().NewGroup("New Test Group")
		if err != nil {
			t.Fatalf("Error creating group for testing")
		}

		group2, err := admin.User().NewGroup("New Test Group2")
		if err != nil {
			t.Fatalf("Error creating group for testing")
		}

		_, err = admin.NewRegistrationToken("test", 0, time.Time{}, []data.ID{group.ID, group2.ID, data.NewID()})
		if !app.IsFail(err) {
			t.Fatalf("Generating a token with at least one invalid groupID did not fail: %s", err)
		}

		token, err := admin.NewRegistrationToken("test", 0, time.Time{}, nil)
		if err != nil {
			t.Fatalf("Generating registration token failed: %s", err)
		}
		if token.Token == "" {
			t.Fatalf("Invalid token: %s", token.Token)
		}
		if token.Description != "test" {
			t.Fatalf("Token description is incorrect. Expected %s, got %s", "test", token.Description)
		}

		if token.Expires.Valid || !token.Expires.Time.IsZero() {
			t.Fatalf("Invalid null expiration value. Expected %v, got %v", time.Time{}, token.Expires)
		}

		if token.Limit != -1 {
			t.Fatalf("Invalid null limit value. Expected %d got %d", -1, token.Limit)
		}

	})

	t.Run("Register User From Token", func(t *testing.T) {
		reset(t)

		_, err := app.RegisterUserFromToken("newuser", "newuserPassword", "GarbageToken")
		if !app.IsFail(err) {
			t.Fatal("Registering user with invalid token did not fail")
		}

		token, err := admin.NewRegistrationToken("test", 0, time.Time{}, nil)
		if err != nil {
			t.Fatalf("Generating registration token failed: %s", err)
		}
		_, err = app.RegisterUserFromToken(admin.User().Username, "newuserPassword", token.Token)
		if !app.IsFail(err) {
			t.Fatal("Registering user with an existing username didn't fail: ", err)
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

		token, err := admin.NewRegistrationToken("test", 3, time.Time{}, nil)
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
		token, err := admin.NewRegistrationToken("test", 0, time.Now().Add(2*time.Second), nil)
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
		g, err := admin.User().NewGroup("Test Group 1")
		if err != nil {
			t.Fatalf("Error adding a new group: %s", err)
		}

		g2, err := admin.User().NewGroup("Test Group 2")
		if err != nil {
			t.Fatalf("Error adding a new group: %s", err)
		}

		token, err := admin.NewRegistrationToken("test", 0, time.Time{}, []data.ID{g.ID, g2.ID})
		if err != nil {
			t.Fatalf("Generating registration token failed: %s", err)
		}

		u, err := app.RegisterUserFromToken("user", "newuserPassword", token.Token)
		if err != nil {
			t.Fatalf("Error registering new user from token: %s", err)
		}

		rows, err := data.NewQuery(`select group_id from group_users where user_id = {{arg "id"}}`).
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

		groups, err := token.Groups()
		if err != nil {
			t.Fatalf("Error getting token groups: %s", err)
		}

		for i := range groups {
			if groups[i].ID != g.ID && groups[i].ID != g2.ID {
				t.Fatalf("Group not found in token's groups. ID: %s ", groups[i].ID)
			}
		}
	})

	t.Run("Users", func(t *testing.T) {
		reset(t)

		token, err := admin.NewRegistrationToken("test", 0, time.Time{}, nil)
		if err != nil {
			t.Fatalf("Generating registration token failed: %s", err)
		}

		u1, err := app.RegisterUserFromToken("firstuser", "newuserPassword", token.Token)
		if err != nil {
			t.Fatalf("Error registering new user from token: %s", err)
		}

		u2, err := app.RegisterUserFromToken("seconduser", "newuserPassword", token.Token)
		if err != nil {
			t.Fatalf("Error registering new user from token: %s", err)
		}

		users, err := token.Users()
		if err != nil {
			t.Fatalf("Error getting users from token: %s", err)
		}

		for i := range users {
			if users[i].ID != u1.ID && users[i].ID != u2.ID {
				t.Fatalf("User not found in token's user list. ID: %s ", users[i].ID)
			}
		}

	})

	t.Run("Valid", func(t *testing.T) {
		reset(t)

		token, err := admin.NewRegistrationToken("test", 0, time.Time{}, nil)
		if err != nil {
			t.Fatalf("Generating registration token failed: %s", err)
		}
		err = token.Invalidate()
		if err != nil {
			t.Fatalf("Error Setting token to invalid: %s", err)
		}

		_, err = app.RegisterUserFromToken("newuser", "newuserpassword", token.Token)
		if !app.IsFail(err) {
			t.Fatalf("Registering a user with an invalid token did not fail: %s", err)
		}

	})

	t.Run("List", func(t *testing.T) {
		reset(t)
		valid := 10
		for i := 0; i < valid; i++ {
			_, err := admin.NewRegistrationToken("test", 0, time.Time{}, nil)
			if err != nil {
				t.Fatalf("Error adding registration tokens: %s", err)
			}
		}

		invalid := 3
		tkn, err := admin.NewRegistrationToken("test", 0, time.Time{}, nil)
		if err != nil {
			t.Fatalf("Error adding registration tokens: %s", err)
		}
		_, err = data.NewQuery(`update registration_tokens set valid = {{FALSE}} where token = {{arg "token"}}`).
			Exec(data.Arg("token", tkn.Token))
		if err != nil {
			t.Fatalf("Error adding registration tokens: %s", err)
		}

		tkn, err = admin.NewRegistrationToken("test", 0, time.Time{}, nil)
		if err != nil {
			t.Fatalf("Error adding registration tokens: %s", err)
		}
		_, err = data.NewQuery(`update registration_tokens set {{limit}} = 0  where token = {{arg "token"}}`).
			Exec(data.Arg("token", tkn.Token))
		if err != nil {
			t.Fatalf("Error invalidating token: %s", err)
		}

		tkn, err = admin.NewRegistrationToken("test", 0, time.Time{}, nil)
		if err != nil {
			t.Fatalf("Error adding registration tokens: %s", err)
		}
		_, err = data.NewQuery(`update registration_tokens set expires = {{arg "expires"}} where token = {{arg "token"}}`).
			Exec(data.Arg("token", tkn.Token), data.Arg("expires", time.Now().Add(-1*time.Hour)))
		if err != nil {
			t.Fatalf("Error invalidating token: %s", err)
		}

		tests := []struct {
			valid  bool
			offset int
			limit  int

			total int
			len   int
		}{
			{false, 0, 20, valid + invalid, valid + invalid},
			{false, 0, 5, valid + invalid, 5},
			{true, 0, 20, valid, valid},
			{true, 0, 5, valid, 5},
			{false, 5, 5, valid + invalid, 5},
			{false, 10, 5, valid + invalid, 3},
			{false, 8, 5, valid + invalid, 5},
			{true, 5, 5, valid, 5},
			{true, 10, 5, valid, 0},
			{true, 8, 5, valid, 2},
			{true, 0, 0, valid, 10},
			{false, 0, 0, valid + invalid, 10},
			{false, 0, 10001, valid + invalid, 10},
		}

		for i, test := range tests {
			t.Run("test-"+strconv.Itoa(i), func(t *testing.T) {
				tokens, total, err := admin.RegistrationTokenList(test.valid, test.offset, test.limit)
				if err != nil {
					t.Fatalf("Error getting registration token list: %s", err)
				}

				if len(tokens) != test.len {
					t.Fatalf("Invalid result length. Expected %d, got %d", test.len, len(tokens))
				}

				if total != test.total {
					t.Fatalf("Expected token list total to be %d, got %d", test.total, total)
				}
				if test.valid {
					for i := range tokens {
						if !tokens[i].Valid() {
							t.Fatalf("Expected all tokens to be valid. This one wasn't: %v",
								tokens[i])
						}
					}
				}
			})
		}
	})

	t.Run("Creator", func(t *testing.T) {
		reset(t)

		token, err := admin.NewRegistrationToken("test", 0, time.Time{}, nil)
		if err != nil {
			t.Fatalf("Generating registration token failed: %s", err)
		}

		creator, err := token.Creator()
		if err != nil {
			t.Fatalf("Error getting registration token creator: %s", err)
		}

		if creator.ID != admin.User().ID {
			t.Fatalf("Registration token creator is incorrect. Expected ID %s, got %s", admin.User().ID,
				creator.ID)
		}

		// test cache

		creator, err = token.Creator()
		if err != nil {
			t.Fatalf("Error getting cached registration token creator: %s", err)
		}

		if creator.ID != admin.User().ID {
			t.Fatalf("Cached Registration token creator is incorrect. Expected ID %s, got %s", admin.User().ID,
				creator.ID)
		}

	})

	t.Run("Get", func(t *testing.T) {
		reset(t)

		token, err := admin.NewRegistrationToken("test", 0, time.Time{}, nil)
		if err != nil {
			t.Fatalf("Generating registration token failed: %s", err)
		}

		other, err := admin.RegistrationToken(token.Token)
		if err != nil {
			t.Fatalf("Error getting registration token: %s", err)
		}

		if other.Token != token.Token {
			t.Fatalf("Retrieved token doesn't match. Expected %s, got %s", token.Token, other.Token)
		}

	})

	t.Run("URL", func(t *testing.T) {
		reset(t)
		uri := "http://testurl.com"
		err := admin.SetSetting("URL", uri)
		if err != nil {
			t.Fatalf("Error setting URL setting: %s", err)
		}

		token, err := admin.NewRegistrationToken("test", 0, time.Time{}, nil)
		if err != nil {
			t.Fatalf("Generating registration token failed: %s", err)
		}

		tokenURL := fmt.Sprintf("%s%s/%s", uri, app.RegistrationTokenPath, token.Token)

		otherURL, err := token.URL()
		if err != nil {
			t.Fatalf("Error getting token URL: %s", err)
		}

		if tokenURL != otherURL {
			t.Fatalf("Token URL is incorrect. Expected %s, got %s", tokenURL, otherURL)
		}
	})
}
