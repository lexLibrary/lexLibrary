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

	serveImage(w, r, u.ProfileImage())
}

func (p *profilePage) loadData(s *app.Session) error {
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
		err := p.loadData(c.session)
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
		err := p.loadData(c.session)
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
		err := p.loadData(c.session)
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
		err := p.loadData(c.session)
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

func profileEditTemplate(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r) {
		return
	}
	err = w.(*templateWriter).execute(u)

	if err != nil {
		app.LogError(errors.Wrap(err, "Executing profile_edit template: %s"))
	}

}
