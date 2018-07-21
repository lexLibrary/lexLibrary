// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
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

// type assertRowTest struct {
// 	query   string
// 	args    []data.Argument
// 	results []interface{}
// }

// type assertRowTests []assertRowTest

// func (a assertRowTests) run(t *testing.T) {
// 	t.Helper()
// 	for _, test := range a {
// 		qry := data.NewQuery(test.query)
// 		err := testRow(qry.QueryRow(test.args...), test.results...)
// 		if err != nil {
// 			t.Fatalf("Error running test query: \n%s\nERROR: %s", qry.Statement(), err)
// 		}
// 	}
// }

// Thanks Ben Johnson https://github.com/benbjohnson/testing

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

// assertFail asserts that an error is returned and is of type failure, and that the http status matches
func assertFail(tb testing.TB, err error, status int, msg string) {
	_, file, line, _ := runtime.Caller(1)
	if !app.IsFail(err) {
		fmt.Printf("\033[31m%s:%d:\n\n\texp failure\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, err)
		tb.FailNow()
	}

	fail := err.(*app.Fail)
	if fail.HTTPStatus != status {
		fmt.Printf("\033[31m%s:%d:\n\n\texp status: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, status,
			fail.HTTPStatus)
		tb.FailNow()
	}
}
