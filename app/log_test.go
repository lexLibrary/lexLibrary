// Copyright (c) 2017 Townsourced Inc.

package app_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
	"github.com/pkg/errors"
)

func TestLog(t *testing.T) {
	_, err := data.NewQuery("delete from logs").Exec()
	if err != nil {
		t.Fatalf("Error emptying logs table before running tests: %s", err)
	}

	t.Run("Log Error", func(t *testing.T) {
		testErr := fmt.Errorf("New test error")

		app.LogError(testErr)
	})
	t.Run("Log Get", func(t *testing.T) {
		for i := 0; i < 12; i++ {
			app.LogError(fmt.Errorf("Error %d", i))
		}

		t.Run("Min", func(t *testing.T) {
			logs, err := app.LogGet(0, 0)
			if err != nil {
				t.Fatalf("Error retrieving the minimum number of logs: %s", err)
			}

			if len(logs) != 10 {
				t.Fatalf("Invalid number of logs, wanted %d got %d", 10, len(logs))
			}
		})
		t.Run("Max", func(t *testing.T) {
			logs, err := app.LogGet(0, 10001)
			if err != nil {
				t.Fatalf("Error retrieving the max number of logs: %s", err)
			}

			if len(logs) != 10 {
				t.Fatalf("Invalid number of logs, wanted %d got %d", 10, len(logs))
			}
		})
		t.Run("First Five", func(t *testing.T) {
			logs, err := app.LogGet(0, 5)
			if err != nil {
				t.Fatalf("Error retrieving first five logs: %s", err)
			}
			if len(logs) != 5 {
				t.Fatalf("Invalid number of logs. Wanted %d got %d", 5, len(logs))
			}
		})
		t.Run("Second Five", func(t *testing.T) {
			logs, err := app.LogGet(5, 5)
			if err != nil {
				t.Fatalf("Error retrieving second five logs: %s", err)
			}

			if len(logs) != 5 {
				t.Fatalf("Invalid number of logs. Wanted %d got %d", 5, len(logs))
			}
		})
		t.Run("Third Five", func(t *testing.T) {
			logs, err := app.LogGet(10, 5)
			if err != nil {
				t.Fatalf("Error retrieving third five logs: %s", err)
			}

			if len(logs) != 3 {
				t.Fatalf("Invalid number of logs. Wanted %d got %d", 3, len(logs))
			}
		})
	})

	t.Run("Search", func(t *testing.T) {
		search := "Search Test"
		app.LogError(fmt.Errorf("Error message with %s", search))
		logs, err := app.LogSearch(strings.ToLower(search), 0, 10)
		if err != nil {
			t.Fatalf("Error searching logs: %s", err)
		}

		if len(logs) != 1 {
			t.Fatalf("Invalid number of logs. Wanted %d got %d", 1, len(logs))
		}

		if !strings.Contains(logs[0].Message, search) {
			t.Fatalf("Log message '%s' does not contain the search value of '%s'", logs[0].Message, search)
		}
	})

	t.Run("Logger", func(t *testing.T) {
		entry := "Test Logger entry"
		logger := app.Logger("Test Prefix: ")

		logger.Print(entry)

		logs, err := app.LogGet(0, 1)
		if err != nil {
			t.Fatalf("Error getting logs after logger test")
		}
		if !strings.Contains(logs[0].Message, entry) {
			t.Fatalf("Incorrect logger entry found. Expected log to contain %s, log was %s",
				entry, logs[0].Message)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		test := errors.New("New error")
		id := app.LogError(test)

		log, err := app.LogGetByID(id)
		if err != nil {
			t.Fatalf("Error getting log by ID: %s", err)
		}

		if log.ID != id || log.Message != test.Error() {
			t.Fatalf("Logged error doesn't match error")
		}

	})
}
