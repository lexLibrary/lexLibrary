// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"net/http"
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
)

func TestFail(t *testing.T) {
	t.Run("New Failure", func(t *testing.T) {
		err := app.NewFailure("New Failure")

		if !app.IsFail(err) {
			t.Fatalf("Error is not a failure")
		}

		if err.(*app.Fail).HTTPStatus != http.StatusBadRequest {
			t.Fatalf("Invalid Status code for general failure.  Expected %d, got %d", http.StatusBadRequest,
				err.(*app.Fail).HTTPStatus)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		err := app.NotFound("Not Found Test")
		fail, ok := err.(*app.Fail)
		if !ok {
			t.Fatalf("Error is not a failure")
		}

		if fail.HTTPStatus != http.StatusNotFound {
			t.Fatalf("Invalid Status code for Not Found.  Expected %d, got %d", http.StatusNotFound, fail.HTTPStatus)
		}
	})
	t.Run("Unauthorized", func(t *testing.T) {
		err := app.Unauthorized("Unauthorized Test")
		fail, ok := err.(*app.Fail)
		if !ok {
			t.Fatalf("Error is not a failure")
		}

		if fail.HTTPStatus != http.StatusUnauthorized {
			t.Fatalf("Invalid Status code for Not Found.  Expected %d, got %d", http.StatusUnauthorized, fail.HTTPStatus)
		}
	})
}
