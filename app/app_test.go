// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
)

func TestMain(m *testing.M) {
	err := data.TestingSetup()
	if err != nil {
		log.Fatal(err)
	}

	result := m.Run()
	err = data.Teardown()
	if err != nil {
		log.Fatalf("Error tearing down data connections: %s", err)
	}
	os.Exit(result)
}

func resetDB(t *testing.T) {
	t.Helper()
	truncateTable(t, "logs")
	truncateTable(t, "sessions")
	truncateTable(t, "user_to_groups")
	truncateTable(t, "registration_token_groups")
	truncateTable(t, "registration_token_users")
	truncateTable(t, "users")
	truncateTable(t, "groups")
	truncateTable(t, "settings")
	truncateTable(t, "images")
	truncateTable(t, "registration_tokens")
}

func resetAdmin(t *testing.T, username, password string) *app.Admin {
	t.Helper()
	resetDB(t)

	u, err := app.FirstRunSetup(username, password)
	if err != nil {
		t.Fatalf("Error setting up admin user: %s", err)
	}
	admin, err := u.Admin()
	if err != nil {
		t.Fatal(err)
	}

	return admin
}

func truncateTable(t *testing.T, table string) {
	t.Helper()
	_, err := data.NewQuery(fmt.Sprintf("delete from %s", table)).Exec()
	if err != nil {
		t.Fatalf("Error emptying %s table before running tests: %s", table, err)
	}
}
