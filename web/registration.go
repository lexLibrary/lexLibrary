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

	Valid bool
}

func registrationCreate(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r) {
		return
	}
	admin, err := u.Admin()
	if errHandled(err, w, r) {
		return
	}

	input := &registrationInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r) {
		return
	}

	token, err := admin.NewRegistrationToken(input.Description, input.Limit, input.Expires, input.Groups)
	if errHandled(err, w, r) {
		return
	}

	respond(w, created(token))
}

func registrationUpdate(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}
}
