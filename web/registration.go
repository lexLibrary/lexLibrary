// Copyright (c) 2017-2018 Townsourced Inc.

package web

import (
	"net/http"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
)

type registrationInput struct {
	Description string
	Limit       uint
	Expires     time.Time
	Groups      []data.ID
}

func registrationTemplate(w http.ResponseWriter, r *http.Request, c ctx) {
	w.(*templateWriter).execute(struct {
		Token string
	}{
		Token: c.params.ByName("token"),
	})
}

func registrationCreate(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	a, err := c.session.Admin()
	if errHandled(err, w, r) {
		return
	}

	input := &registrationInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r) {
		return
	}

	token, err := a.NewRegistrationToken(input.Description, input.Limit, input.Expires, input.Groups)
	if errHandled(err, w, r) {
		return
	}

	respond(w, created(token))
}

func registrationDelete(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	a, err := c.session.Admin()
	if errHandled(err, w, r) {
		return
	}

	token, err := a.RegistrationToken(c.params.ByName("token"))
	if errHandled(err, w, r) {
		return
	}

	if errHandled(token.Invalidate(), w, r) {
		return
	}

	respond(w, success(nil))
}
