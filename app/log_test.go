// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/pkg/errors"
)

func TestLog(t *testing.T) {

	truncateTable(t, "logs")

	t.Run("Log Error", func(t *testing.T) {
		testErr := fmt.Errorf("New test error")

		app.LogError(testErr)
	})
	t.Run("Log Get", func(t *testing.T) {
		truncateTable(t, "logs")
		count := 12
		for i := 0; i < count; i++ {
			app.LogError(fmt.Errorf("Error %d", i))
		}

		tests := []struct {
			name   string
			offset int
			limit  int

			len   int
			total int
		}{
			{"Min", 0, 0, 10, count},
			{"Max", 0, 10001, 10, count},
			{"First Five", 0, 5, 5, count},
			{"Second Five", 5, 5, 5, count},
			{"Third Five", 10, 5, 2, count},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				logs, total, err := app.LogGet(test.offset, test.limit)
				if err != nil {
					t.Fatalf("Error retrieving logs: %s", err)
				}

				if len(logs) != test.len {
					t.Fatalf("Invalid number of logs, wanted %d got %d", test.len, len(logs))
				}
				if total != test.total {
					t.Fatalf("Invalid total number of logs. Wanted %d got %d", test.total, total)
				}
			})
		}

	})

	t.Run("Search", func(t *testing.T) {
		search := "Search Test"
		app.LogError(fmt.Errorf("Error message with %s", search))
		logs, total, err := app.LogSearch(strings.ToLower(search), 0, 10)
		if err != nil {
			t.Fatalf("Error searching logs: %s", err)
		}

		if !strings.Contains(logs[0].Message, search) {
			t.Fatalf("Log message '%s' does not contain the search value of '%s'", logs[0].Message, search)
		}
		if total != 1 {
			t.Fatalf("Invalid search total. Expected %d, got %d", 1, total)
		}

	})

	t.Run("Logger", func(t *testing.T) {
		entry := "Test Logger entry"
		logger := app.Logger("Test Prefix: ")

		logger.Print(entry)

		logs, _, err := app.LogGet(0, 1)
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
