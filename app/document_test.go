// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"testing"

	"github.com/lexLibrary/lexLibrary/app"
)

func TestDocument(t *testing.T) {
	var admin *app.Admin
	reset := func(t *testing.T) {
		t.Helper()

		admin = resetAdmin(t, "admin", "adminpassword")
	}

}
