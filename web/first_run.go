// Copyright (c) 2017-2018 Townsourced Inc.

package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/lexLibrary/lexLibrary/app"
)

func init() {
	app.FirstRunTrigger(func() {
		addInterrupt(firstRunHandler.ServeHTTP)
	})
}

var firstRunHandler = templateHandler{
	handler:       firstRunTemplate,
	templateFiles: []string{"first_run.template.html"},
}

type firstRunInput struct {
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	Settings map[string]interface{}
}

func firstRunTemplate(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if r.Method == "GET" {
		errHandled(w.(*templateWriter).execute(struct {
		}{}), w, r)
		return
	}

	if r.Method != "POST" {
		notFound(w, r)
		return
	}

	input := &firstRunInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r) {
		return
	}

	if input.Username == nil {
		errHandled(app.NewFailure("A username is required"), w, r)
		return
	}

	empty := ""
	if input.Password == nil {
		input.Password = &empty
	}

	u, err := app.FirstRunSetup(*input.Username, *input.Password, input.Settings)
	if errHandled(err, w, r) {
		return
	}

	setSession(w, r, u, false)
	removeInterrupt(firstRunHandler.ServeHTTP)
}
