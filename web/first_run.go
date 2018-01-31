// Copyright (c) 2017-2018 Townsourced Inc.

package web

import (
	"net/http"

	"github.com/lexLibrary/lexLibrary/app"
)

func init() {
	firstRunInterrupt = &interrupt{
		name: "firstRun",
		fn:   firstRunHandler,
	}

	app.FirstRunTrigger(func() {
		addInterrupt(firstRunInterrupt)
	})

	firstRunTemplate.loadTemplates()
}

var firstRunInterrupt *interrupt
var firstRunTemplate = templateHandler{
	templateFiles: []string{"first_run.template.html"},
}

func firstRunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		errHandled(firstRunTemplate.template.Execute(w, nil), w, r)
		return
	}

	if r.Method != "POST" {
		notFound(w, r)
		return
	}

	input := &userInput{}
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

	u, err := app.FirstRunSetup(*input.Username, *input.Password)
	if errHandled(err, w, r) {
		return
	}
	removeInterrupt(firstRunInterrupt)

	setSession(w, r, u, false)
}
