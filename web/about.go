// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/pkg/errors"
)

func aboutTemplate(w http.ResponseWriter, r *http.Request, c ctx) {
	var u *app.User
	if c.session != nil {

	}
	err := w.(*templateWriter).execute(struct {
		Version   string
		BuildDate string
		Runtime   *app.RuntimeInfo
	}{
		Version:   app.Version(),
		BuildDate: app.BuildDate().Format(time.Stamp),
		Runtime:   app.Runtime(u),
	})

	if err != nil {
		app.LogError(errors.Wrap(err, "Executing about template: %s"))
	}
}
