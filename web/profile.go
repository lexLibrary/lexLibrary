// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/lexLibrary/lexLibrary/app"
	"github.com/pkg/errors"
)

type profilePage struct {
	templateHandler
	data struct {
		User  *app.User
		Stats app.UserStats
		Tab   string
	}
}

func profileGetImage(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r) {
		return
	}

	if _, ok := r.URL.Query()["draft"]; ok {
		serveImage(w, r, u.ProfileImageDraft())
		return
	}

	serveImage(w, r, u.ProfileImage())
}

func (p *profilePage) loadShared(s *app.Session) error {
	if s == nil {
		return app.Unauthorized("You do not have access to this user")
	}

	u, err := s.User()
	if err != nil {
		return err
	}

	stats, err := u.Stats()
	if err != nil {
		return err
	}

	p.data.User = u
	p.data.Stats = stats
	return nil
}

func (p *profilePage) documents(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {

	p.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		err := p.loadShared(c.session)
		if errHandled(err, w, r) {
			return
		}

		p.data.Tab = "documents"
		err = w.(*templateWriter).execute(p.data)

		if err != nil {
			app.LogError(errors.Wrap(err, "Executing profile template: %s"))
		}
	}
	p.ServeHTTP(w, r, parms)
}

func (p *profilePage) readLater(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {

	p.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		err := p.loadShared(c.session)
		if errHandled(err, w, r) {
			return
		}

		p.data.Tab = "readLater"
		err = w.(*templateWriter).execute(p.data)

		if err != nil {
			app.LogError(errors.Wrap(err, "Executing profile template: %s"))
		}
	}
	p.ServeHTTP(w, r, parms)

}

func (p *profilePage) comments(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {

	p.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		err := p.loadShared(c.session)
		if errHandled(err, w, r) {
			return
		}

		p.data.Tab = "comments"
		err = w.(*templateWriter).execute(p.data)

		if err != nil {
			app.LogError(errors.Wrap(err, "Executing profile template: %s"))
		}
	}
	p.ServeHTTP(w, r, parms)

}

func (p *profilePage) history(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {

	p.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		err := p.loadShared(c.session)
		if errHandled(err, w, r) {
			return
		}

		p.data.Tab = "history"
		err = w.(*templateWriter).execute(p.data)

		if err != nil {
			app.LogError(errors.Wrap(err, "Executing profile template: %s"))
		}
	}
	p.ServeHTTP(w, r, parms)

}

type profileEditPage struct {
	templateHandler
	data struct {
		User *app.User
		Tab  string
	}
}

func (p *profileEditPage) loadShared(s *app.Session) error {
	if s == nil {
		return app.Unauthorized("You do not have access to this user")
	}

	u, err := s.User()
	if err != nil {
		return err
	}

	p.data.User = u
	return nil
}

func (p *profileEditPage) root(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {
	p.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		err := p.loadShared(c.session)
		if errHandled(err, w, r) {
			return
		}
		p.data.Tab = "profile"

		err = w.(*templateWriter).execute(p.data)

		if err != nil {
			app.LogError(errors.Wrap(err, "Executing profile template: %s"))
		}
	}
	p.ServeHTTP(w, r, parms)
}

func (p *profileEditPage) account(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {
	p.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		err := p.loadShared(c.session)
		if errHandled(err, w, r) {
			return
		}
		p.data.Tab = "account"

		err = w.(*templateWriter).execute(p.data)

		if err != nil {
			app.LogError(errors.Wrap(err, "Executing profile template: %s"))
		}
	}
	p.ServeHTTP(w, r, parms)
}

func profileUpdateName(w http.ResponseWriter, r *http.Request, c ctx) {
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

func profileUpdateUsername(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	input := &userInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r) {
		return
	}

	if input.Username == nil {
		errHandled(app.NewFailure("username is required"), w, r)
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

	if errHandled(u.SetUsername(*input.Username, *input.Version), w, r) {
		return
	}

	respond(w, success(nil))
}

func profileUploadImage(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r) {
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

	if errHandled(u.UploadProfileImageDraft(uploads[0], u.Version), w, r) {
		return
	}

	respond(w, created(nil))

}

func profileCropImage(w http.ResponseWriter, r *http.Request, c ctx) {
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

	if errHandled(u.SetProfileImageFromDraft(input.X0, input.Y0, input.X1, input.Y1), w, r) {
		return
	}

	respond(w, success(nil))
}
