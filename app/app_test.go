// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"fmt"
	"log"
	"os"
	"testing"

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

func truncateTable(t *testing.T, table string) {
	_, err := data.NewQuery(fmt.Sprintf("delete from %s", table)).Exec()
	if err != nil {
		t.Fatalf("Error emptying %s table before running tests: %s", table, err)
	}
}
