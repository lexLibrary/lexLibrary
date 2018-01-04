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
	reset := func() {
		t.Helper()
		_, err := data.NewQuery("delete from sessions").Exec()
		if err != nil {
			t.Fatalf("Error emptying sessions table before running tests: %s", err)
		}

	}

	u, err := app.UserNew("testusername", "", "", "ODSjflaksjdfhiasfd323")
	if err != nil {
		t.Fatalf("Error adding user for session testing")
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

}
