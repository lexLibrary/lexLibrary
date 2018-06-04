// Copyright (c) 2017-2018 Townsourced Inc.
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
	Name     *string `json:"name,omitempty"`
	Version  *int    `json:"version,omitempty"`
}

type passwordInput struct {
	Username    *string `json:"username,omitempty"`
	OldPassword *string `json:"oldPassword,omitempty"`
	NewPassword *string `json:"newPassword,omitempty"`
	Version     *int    `json:"version,omitempty"`
}

type userImageInput struct {
	X0 float64 `json:"x0"`
	Y0 float64 `json:"y0"`
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
}

func userCreate(w http.ResponseWriter, r *http.Request, c ctx) {
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
	token := c.params.ByName("token")

	var u *app.User

	if token == "" {
		u, err = app.UserNew(*input.Username, *input.Password)
	} else {
		u, err = app.RegisterUserFromToken(*input.Username, *input.Password, token)
	}
	if errHandled(err, w, r) {
		return
	}
	_, err = setSession(w, r, u, false)
	if errHandled(err, w, r) {
		return
	}

	respond(w, created(u))
}

func usernameTest(w http.ResponseWriter, r *http.Request, c ctx) {
	username := c.params.ByName("username")

	_, err := app.UserGet(username)
	if errHandled(err, w, r) {
		return
	}

	respond(w, success(nil))
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

func userUpdatePassword(w http.ResponseWriter, r *http.Request, c ctx) {
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

		err = u.SetPassword(*input.OldPassword, *input.NewPassword, *input.Version)
		if errHandled(err, w, r) {
			return
		}
		expireSessionCookie(w, r, c.session)

		_, err = setSession(w, r, u, false)
		if errHandled(err, w, r) {
			return
		}

		respond(w, success(nil))
		return
	}

	if input.Username == nil {
		notFound(w, r)
		return
	}

	// user's password has expired and being set from login prompt
	u, err := app.UserSetExpiredPassword(*input.Username, *input.OldPassword, *input.NewPassword)
	if errHandled(err, w, r) {
		return
	}

	_, err = setSession(w, r, u, false)
	if errHandled(err, w, r) {
		return
	}

	respond(w, success(nil))
}

func userGetImage(w http.ResponseWriter, r *http.Request, c ctx) {
	username := c.params.ByName("username")

	u, err := app.UserGet(username)
	if errHandled(err, w, r) {
		return
	}

	serveImage(w, r, u.ProfileImage(), false)
}
