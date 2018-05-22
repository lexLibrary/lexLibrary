// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"

	"github.com/lexLibrary/lexLibrary/app"
)

func rootTemplate(w http.ResponseWriter, r *http.Request, c ctx) {
	var u *app.User
	var err error
	if c.session != nil {
		u, err = c.session.User()
		if errHandled(err, w, r) {
			return
		}
	}

	w.(*templateWriter).execute(struct {
		User *app.User
	}{
		User: u,
	})
}
