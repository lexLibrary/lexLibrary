// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/pkg/errors"
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

		if u.Username != strings.ToLower(c.params.ByName("username")) {
			unauthorized(w, r)
			return
		}

		err = u.SetPassword(*input.OldPassword, *input.NewPassword, *input.Version)
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

func userUpdateName(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	input := &userInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r) {
		return
	}

	if input.Name == nil {
		errHandled(app.NewFailure("name is required"), w, r)
		return
	}

	if input.Version == nil {
		errHandled(app.NewFailure("version is required"), w, r)
		return
	}
	u, err := c.session.User()
	if errHandled(err, w, r) {
		return
	}

	if errHandled(u.SetName(*input.Name, *input.Version), w, r) {
		return
	}

	respond(w, success(nil))

}

func userUploadImage(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	input := &userInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r) {
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r) {
		return
	}

	if input.Version == nil {
		errHandled(app.NewFailure("version is required"), w, r)
		return
	}

	uploads, err := filesFromForm(r)
	if errHandled(err, w, r) {
		return
	}
	if len(uploads) == 0 {
		errHandled(app.NewFailure("No image uploaded"), w, r)
		return
	}

	if len(uploads) > 1 {
		errHandled(app.NewFailure("More than one image was uploaded.  Please upload one image at a time"), w, r)
		return
	}

	if errHandled(u.SetProfileImage(uploads[0], *input.Version), w, r) {
		return
	}

	respond(w, created(nil))

}

func userCropImage(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	input := &userImageInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r) {
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r) {
		return
	}

	if errHandled(u.CropProfileImage(input.X0, input.Y0, input.X1, input.Y1), w, r) {
		return
	}

	respond(w, success(nil))
}

type profileTemplateData struct {
	User  *app.User
	Stats app.UserStats
	Tab   string
}

func profileTemplate(w http.ResponseWriter, r *http.Request, c ctx) {
	data, ok := profileView(w, r, c)
	if !ok {
		return
	}

	data.Tab = "documents"

	err := w.(*templateWriter).execute(data)

	if err != nil {
		app.LogError(errors.Wrap(err, "Executing profile template: %s"))
	}
}

// func profileTemplateDocuments(w http.ResponseWriter, r *http.Request, c ctx) {
// 	data, ok := profileView(w, r, c)
// 	if !ok {
// 		return
// 	}

// 	data.Tab = "documents"

// 	err := w.(*templateWriter).execute(data)

// 	if err != nil {
// 		app.LogError(errors.Wrap(err, "Executing profile template: %s"))
// 	}
// }

// func profileTemplateReadLater(w http.ResponseWriter, r *http.Request, c ctx) {
// 	data, ok := profileView(w, r, c)
// 	if !ok {
// 		return
// 	}

// 	data.Tab = "readLater"

// 	err := w.(*templateWriter).execute(data)

// 	if err != nil {
// 		app.LogError(errors.Wrap(err, "Executing profile template: %s"))
// 	}
// }

// func profileTemplateComments(w http.ResponseWriter, r *http.Request, c ctx) {
// 	data, ok := profileView(w, r, c)
// 	if !ok {
// 		return
// 	}

// 	data.Tab = "comments"

// 	err := w.(*templateWriter).execute(data)

// 	if err != nil {
// 		app.LogError(errors.Wrap(err, "Executing profile template: %s"))
// 	}
// }

// func profileTemplateHistory(w http.ResponseWriter, r *http.Request, c ctx) {
// 	data, ok := profileView(w, r, c)
// 	if !ok {
// 		return
// 	}

// 	data.Tab = "history"

// 	err := w.(*templateWriter).execute(data)

// 	if err != nil {
// 		app.LogError(errors.Wrap(err, "Executing profile template: %s"))
// 	}
// }

func profileView(w http.ResponseWriter, r *http.Request, c ctx) (*profileTemplateData, bool) {
	if c.session == nil {
		unauthorized(w, r)
		return nil, false
	}

	u, err := c.session.User()
	if errHandled(err, w, r) {
		return nil, false
	}

	stats, err := u.Stats()
	if errHandled(err, w, r) {
		return nil, false
	}

	return &profileTemplateData{
		User:  u,
		Stats: stats,
	}, true
}

func profileEditTemplate(w http.ResponseWriter, r *http.Request, c ctx) {
	unauthorized(w, r)
}
