// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"fmt"
	"log"
	"os"
	"reflect"
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
	truncateTable(t, "document_groups")
	truncateTable(t, "document_tags")
	truncateTable(t, "document_draft_tags")
	truncateTable(t, "document_history")
	truncateTable(t, "document_drafts")
	truncateTable(t, "documents")
	truncateTable(t, "groups")
	truncateTable(t, "settings")
	truncateTable(t, "registration_tokens")
	truncateTable(t, "users")
	truncateTable(t, "images")
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

func assertRow(t *testing.T, row *data.Row, assertValues ...interface{}) {
	t.Helper()
	rowValues := make([]interface{}, len(assertValues), len(assertValues))

	for i := range assertValues {
		rowValues[i] = reflect.New(reflect.TypeOf(assertValues[i])).Interface()
	}

	err := row.Scan(rowValues...)
	if err != nil {
		t.Fatal(err)
		// return err
	}

	for i := range assertValues {
		rowVal := reflect.ValueOf(rowValues[i]).Elem().Interface()
		if !reflect.DeepEqual(assertValues[i], rowVal) {
			t.Fatalf("Column %d doesn't match the asserted value. Expected %v, got %v", i+1,
				assertValues[i], rowVal)
		}
	}
}
