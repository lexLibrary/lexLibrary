// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"testing"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
)

func TestSession(t *testing.T) {
	username := "testusername"
	password := "ODSjflaksjd$hiasfd323"
	var u *app.User
	var admin *app.Admin

	reset := func(t *testing.T) {
		admin = resetAdmin(t, username, password)
		u = admin.User()
	}

	t.Run("New", func(t *testing.T) {
		reset(t)
		expires := time.Time{}
		ipAddress := "127.0.0.1"
		userAgent := "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:57.0) Gecko/20100101 Firefox/57.0"
		s, err := u.NewSession(expires, ipAddress, userAgent)
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
		reset(t)
		s, err := u.NewSession(time.Time{}, "127.0.0.1", "")
		if err != nil {
			t.Fatalf("Error adding new session: %s", err)
		}

		err = s.Logout()
		if err != nil {
			t.Fatalf("Error logging out session: %s", err)
		}

		valid := true
		err = data.NewQuery(`select valid from sessions where id = {{arg "id"}}`).QueryRow(data.Arg("id", s.ID)).
			Scan(&valid)
		if err != nil {
			t.Fatalf("Error getting session for ID %s: %s", s.ID, err)
		}

		if valid {
			t.Fatalf("Logged out session is still valid")
		}

	})

	t.Run("User", func(t *testing.T) {
		reset(t)
		s, err := u.NewSession(time.Time{}, "127.0.0.1", "")
		if err != nil {
			t.Fatalf("Error adding new session: %s", err)
		}

		// get session without cached user
		s, err = app.SessionGet(s.UserID, s.ID)
		if err != nil {
			t.Fatalf("Error getting session: %s", err)
		}

		other, err := s.User()
		if err != nil {
			t.Fatalf("Error getting user from session: %s", err)
		}

		if u.Username != other.Username || u.ID != other.ID {
			t.Fatalf("User from session is not equal to created user. Expected %v, got %v", u, other)
		}

	})
	t.Run("Cycle CSRF", func(t *testing.T) {
		reset(t)
		s, err := u.NewSession(time.Time{}, "127.0.0.1", "")
		if err != nil {
			t.Fatalf("Error adding new session: %s", err)
		}

		// update session's csrf date to more than 15 minutes ago so it'll reset
		_, err = data.NewQuery(`update sessions set csrf_date = {{arg "expires"}} where id = {{arg "id"}}`).Exec(
			data.Arg("expires", time.Now().Add(-20*time.Minute)),
			data.Arg("id", s.ID),
		)
		if err != nil {
			t.Fatalf("Error setting csrf token date: %s", err)
		}

		original := s.CSRFToken

		s, err = app.SessionGet(u.ID, s.ID)
		if err != nil {
			t.Fatalf("Error getting session: %s", err)
		}

		err = s.CycleCSRF()
		if err != nil {
			t.Fatalf("Error resetting csrf token in session: %s", err)
		}

		token := ""
		err = data.NewQuery(`select csrf_token from sessions where id = {{arg "id"}}`).QueryRow(data.Arg("id", s.ID)).
			Scan(&token)
		if err != nil {
			t.Fatalf("Error getting session for ID %s: %s", s.ID, err)
		}

		if original == token || token == "" {
			t.Fatalf("CSRF token was not updated properly: %s", token)
		}

	})

	t.Run("Login", func(t *testing.T) {
		reset(t)
		var err error
		u, err = app.Login(username, password)
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

		err = admin.SetUserActive(u, false, u.Version)
		if err != nil {
			t.Fatalf("Error inactivating user: %s", err)
		}

		_, err = app.Login(username, password)
		if err != app.ErrLogonFailure {
			t.Fatalf("Logging in with inactive user was not a login failure: %v", err)
		}

	})

	t.Run("Get", func(t *testing.T) {
		reset(t)
		s, err := u.NewSession(time.Now().AddDate(0, 0, 1), "", "")
		if err != nil {
			t.Fatalf("Error adding new session: %s", err)
		}

		other, err := app.SessionGet(u.ID, s.ID)

		if err != nil {
			t.Fatalf("Error getting valid session: %s", err)
		}

		if s.ID != other.ID ||
			s.UserID != other.UserID ||
			s.CSRFToken != other.CSRFToken ||
			s.Valid != other.Valid {
			t.Fatalf("Retrieved session does not match. Expected %v got %v", s, other)
		}

		err = s.Logout()
		if err != nil {
			t.Fatalf("Error logging out session")
		}
		_, err = app.SessionGet(u.ID, s.ID)

		if err != app.ErrSessionInvalid {
			t.Fatalf("Session is not invalid when logged out")
		}

	})

	t.Run("Login with Expired Password", func(t *testing.T) {
		reset(t)

		_, err := data.NewQuery(`update users set password_expiration = {{arg "expire"}} where id = {{arg "id"}}`).
			Exec(data.Arg("expire", time.Now().AddDate(0, 0, -1)), data.Arg("id", u.ID))
		if err != nil {
			t.Fatalf("Error expiring user's password: %s", err)
		}

		_, err = app.Login(username, password)
		if err != app.ErrPasswordExpired {
			t.Fatalf("Logging in with an expired password did not result in correct error. Wanted %s, got %s",
				app.ErrPasswordExpired, err)
		}

	})
	t.Run("Admin", func(t *testing.T) {
		reset(t)

		s, err := u.NewSession(time.Time{}, "", "")
		if err != nil {
			t.Fatalf("Error adding new session: %s", err)
		}

		_, err = s.Admin()
		if err != nil {
			t.Fatalf("Error getting admin from session: %s", err)
		}
	})
}
