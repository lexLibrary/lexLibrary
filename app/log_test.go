// Copyright (c) 2017 Townsourced Inc.

package app_test

import (
	"testing"

	"fmt"

	"github.com/lexLibrary/lexLibrary/app"
)

func TestLog(t *testing.T) {
	err := initApp()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Log Error", func(t *testing.T) {
		testErr := fmt.Errorf("New test error")

		app.LogError(testErr)

	})

}
