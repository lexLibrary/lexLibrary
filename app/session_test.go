// Copyright (c) 2017 Townsourced Inc.

package app_test

import (
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
	if err == nil {
		t.Fatalf("No error adding user for session testing")
	}

	t.Run("New", func(t *testing.T) {
		reset()
		expires := time.Now().AddDate(0, 0, 10).Round(time.Second)
		ipAddress := "127.0.0.1"
		userAgent := "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:57.0) Gecko/20100101 Firefox/57.0"
		s, err := app.SessionNew(u, expires, ipAddress, userAgent)
		if err != nil {
			t.Fatalf("Error adding new session: %s", err)
		}

		if s.UserID != u.ID {
			t.Fatalf("Invalid userID in session. Expected %s, got %s", u.ID, s.UserID)
		}
		if !s.Expires.Equal(expires) {
			t.Fatalf("Invalid expires in session. Expected %s, got %s", expires, s.Expires)
		}

	})
}
