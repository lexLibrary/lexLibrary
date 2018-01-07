// Copyright (c) 2017 Townsourced Inc.

package app_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
)

func TestSession(t *testing.T) {
	username := "testusername"
	password := "ODSjflaksjd$hiasfd323"
	var u *app.User

	reset := func() {
		t.Helper()
		_, err := data.NewQuery("delete from sessions").Exec()
		if err != nil {
			t.Fatalf("Error emptying sessions table before running tests: %s", err)
		}
		_, err = data.NewQuery("delete from users").Exec()
		if err != nil {
			t.Fatalf("Error emptying users table before running tests: %s", err)
		}

		u, err = app.UserNew(username, "", "", password)
		if err != nil {
			t.Fatalf("Error adding user for session testing")
		}
	}

	t.Run("New", func(t *testing.T) {
		reset()
		expires := time.Time{}
		ipAddress := "127.0.0.1"
		userAgent := "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:57.0) Gecko/20100101 Firefox/57.0"
		s, err := app.SessionNew(u, expires, ipAddress, userAgent)
		if err != nil {
			t.Fatalf("Error adding new session: %s", err)
		}

		if s.UserID != u.ID {
			t.Fatalf("Invalid userID in session. Expected %s, got %s", u.ID, s.UserID)
		}
		if s.Expires.IsZero() {
			t.Fatalf("Invalid expires in session. %s", s.Expires)
		}
		if s.IPAddress != ipAddress {
			t.Fatalf("Invalid IP address. Expected %s, got %s", ipAddress, s.IPAddress)
		}

		if s.UserAgent != userAgent {
			t.Fatalf("Invalid user agent. Expected %s got %s", userAgent, s.UserAgent)
		}
	})

	t.Run("Logout", func(t *testing.T) {
		reset()
		s, err := app.SessionNew(u, time.Time{}, "127.0.0.1", "")
		if err != nil {
			t.Fatalf("Error adding new session: %s", err)
		}

		err = s.Logout()
		if err != nil {
			t.Fatalf("Error logging out session: %s", err)
		}

		valid := true
		err = data.NewQuery(`select valid from sessions where id = {{arg "id"}}`).QueryRow(sql.Named("id", s.ID)).
			Scan(&valid)
		if err != nil {
			t.Fatalf("Error getting session for ID %s: %s", s.ID, err)
		}

		if valid {
			t.Fatalf("Logged out session is still valid")
		}

	})

	t.Run("User", func(t *testing.T) {
		reset()
		s, err := app.SessionNew(u, time.Time{}, "127.0.0.1", "")
		if err != nil {
			t.Fatalf("Error adding new session: %s", err)
		}

		other, err := s.User()
		if err != nil {
			t.Fatalf("Error getting user from session: %s", err)
		}

		if u.Username != other.Username || u.ID != other.ID {
			t.Fatalf("User from session is not equal to created user. Expected %v, got %v", u, other)
		}

	})
	t.Run("Reset CSRF", func(t *testing.T) {
		reset()
		s, err := app.SessionNew(u, time.Time{}, "127.0.0.1", "")
		if err != nil {
			t.Fatalf("Error adding new session: %s", err)
		}

		original := s.CSRFToken
		err = s.ResetCSRF()
		if err != nil {
			t.Fatalf("Error resetting csrf token in session: %s", err)
		}

		token := ""
		err = data.NewQuery(`select csrf_token from sessions where id = {{arg "id"}}`).QueryRow(sql.Named("id", s.ID)).
			Scan(&token)
		if err != nil {
			t.Fatalf("Error getting session for ID %s: %s", s.ID, err)
		}

		if original == token || token == "" {
			t.Fatalf("CSRF token was not updated properly: %s", token)
		}

	})

	t.Run("Login", func(t *testing.T) {
		reset()
		_, err := app.Login(username, password)
		if err != nil {
			t.Fatalf("Error logging in: %s", err)
		}

		_, err = app.Login("badusername", password)
		if err != app.ErrLogonFailure {
			t.Fatalf("Logging in with bad username was not a login failure: %s", err)
		}

		_, err = app.Login(username, password+"bad")
		if err != app.ErrLogonFailure {
			t.Fatalf("Logging in with bad password was not a login failure: %s", err)
		}

		_, err = app.Login("", password)
		if err != app.ErrLogonFailure {
			t.Fatalf("Logging in with empty username was not a login failure: %s", err)
		}

		_, err = app.Login(username, "1")
		if err != app.ErrLogonFailure {
			t.Fatalf("Logging in with short password was not a login failure: %s", err)
		}

		err = u.SetActive(false, u)
		if err != nil {
			t.Fatalf("Error inactivating user: %s", err)
		}

		_, err = app.Login(username, password)
		if err != app.ErrLogonFailure {
			t.Fatalf("Logging in with inactive user was not a login failure: %s", err)
		}

	})

	t.Run("Get", func(t *testing.T) {
		reset()
		s, err := app.SessionNew(u, time.Now().AddDate(0, 0, 1), "", "")
		if err != nil {
			t.Fatalf("Error adding new session: %s", err)
		}

		other, err := app.SessionGet(s.ID, u.ID)

		if err != nil {
			t.Fatalf("Error getting valid session: %s", err)
		}

		if s.ID != other.ID ||
			s.UserID != other.UserID ||
			!s.Expires.Round(time.Second).Equal(other.Expires.Round(time.Second)) ||
			s.CSRFToken != other.CSRFToken ||
			s.Valid != other.Valid {
			t.Fatalf("Retrieved session does not match. Expected %v got %v", s, other)
		}

		err = s.Logout()
		if err != nil {
			t.Fatalf("Error logging out session")
		}
		_, err = app.SessionGet(s.ID, u.ID)

		if err != app.ErrSessionInvalid {
			t.Fatalf("Session isnot invalid when logged out")
		}

	})
}
