// Copyright (c) 2017 Townsourced Inc.
package web

import (
	"net/http"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/pkg/errors"
)

func rootTemplate(w http.ResponseWriter, r *http.Request, c ctx) {
	err := w.(*templateWriter).execute(struct {
		Test string
	}{
		Test: "test string",
	})

	if err != nil {
		app.LogError(errors.Wrap(err, "Executing ROOT template: %s"))
	}
}
