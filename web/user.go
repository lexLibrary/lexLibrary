// Copyright (c) 2018 Townsourced Inc.
package web

import (
	"net/http"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
)

var publicUserNewRateDelay = app.RateDelay{
	Type:   "userNew",
	Limit:  2,
	Delay:  15 * time.Second,
	Period: 15 * time.Minute,
	Max:    1 * time.Minute,
}

type userInput struct {
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
}

func userPost(w http.ResponseWriter, r *http.Request, c ctx) {
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

	if c.session == nil {
		// don't let public users create too many users quickly
		if errHandled(publicUserNewRateDelay.Attempt(ipAddress(r)), w, r) {
			return
		}
	}

	u, err := app.UserNew(*input.Username, *input.Password)
	if errHandled(err, w, r) {
		return
	}

	if errHandled(setSessionCookie(w, r, u, false), w, r) {
		return
	}

	respond(w, created(u))
}

func userGet(w http.ResponseWriter, r *http.Request, c ctx) {
	username := c.params.ByName("username")

	var who *app.User
	var err error

	if c.session != nil {
		who, err = c.session.User()
		if errHandled(err, w, r) {
			return
		}
	}

	u, err := app.UserGet(username, who)
	if errHandled(err, w, r) {
		return
	}

	respond(w, success(u))
}
