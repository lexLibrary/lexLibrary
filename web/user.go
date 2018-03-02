// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"
	"strings"
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

type passwordInput struct {
	OldPassword *string `json:"oldPassword,omitempty"`
	NewPassword *string `json:"newPassword,omitempty"`
	Version     *int    `json:"version,omitempty"`
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
	_, err = setSession(w, r, u, false)
	if errHandled(err, w, r) {
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

func passwordTest(w http.ResponseWriter, r *http.Request, c ctx) {
	input := &userInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r) {
		return
	}

	password := ""
	if input.Password != nil {
		password = *input.Password
	}

	err = app.ValidatePassword(password)
	if errHandled(err, w, r) {
		return
	}

	respond(w, success(nil))
}

func userPutPassword(w http.ResponseWriter, r *http.Request, c ctx) {
	input := &passwordInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r) {
		return
	}

	if input.OldPassword == nil {
		errHandled(app.NewFailure("oldPassword is required"), w, r)
		return
	}
	if input.NewPassword == nil {
		errHandled(app.NewFailure("newPassword is required"), w, r)
		return
	}

	if c.session != nil {
		// user changing password while logged in
		if input.Version == nil {
			errHandled(app.NewFailure("version is required"), w, r)
			return
		}
		u, err := c.session.User()
		if errHandled(err, w, r) {
			return
		}

		if u.Username != strings.ToLower(c.params.ByName("username")) {
			unauthorized(w, r)
			return
		}

		err = u.SetPassword(*input.OldPassword, *input.NewPassword, *input.Version, u)
		if errHandled(err, w, r) {
			return
		}

		respond(w, success(nil))
		return
	}

	// user's password has expired and being set from login prompt
	u, err := app.UserSetExpiredPassword(c.params.ByName("username"), *input.OldPassword, *input.NewPassword)
	if errHandled(err, w, r) {
		return
	}

	_, err = setSession(w, r, u, false)
	if errHandled(err, w, r) {
		return
	}

	respond(w, success(nil))
}
