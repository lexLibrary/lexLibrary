// Copyright (c) 2017-2018 Townsourced Inc.

package data_test

import (
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
