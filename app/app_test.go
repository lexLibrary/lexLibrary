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
	err := data.TestingSetup(m)
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

func resetAdmin(t *testing.T, username, password string) *app.Admin {
	t.Helper()
	data.ResetDB(t)

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
	err := testRow(row, assertValues...)
	if err != nil {
		t.Fatal(err)
	}
}

func testRow(row *data.Row, assertValues ...interface{}) error {
	rowValues := make([]interface{}, len(assertValues))

	for i := range assertValues {
		rowValues[i] = reflect.New(reflect.TypeOf(assertValues[i])).Interface()
	}

	err := row.Scan(rowValues...)
	if err != nil {
		return err
	}

	for i := range assertValues {
		rowVal := reflect.ValueOf(rowValues[i]).Elem().Interface()
		if !reflect.DeepEqual(assertValues[i], rowVal) {
			return fmt.Errorf("Column %d doesn't match the asserted value. Expected %v, got %v", i+1,
				assertValues[i], rowVal)
		}
	}

	return nil
}

type assertRowTest struct {
	query   string
	args    []data.Argument
	results []interface{}
}

type assertRowTests []assertRowTest

func (a assertRowTests) run(t *testing.T) {
	t.Helper()
	for _, test := range a {
		qry := data.NewQuery(test.query)
		err := testRow(qry.QueryRow(test.args...), test.results...)
		if err != nil {
			t.Fatalf("Error running test query: \n%s\nERROR: %s", qry.Statement(), err)
		}
	}
}
