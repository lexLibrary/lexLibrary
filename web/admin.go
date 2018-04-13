// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/lexLibrary/lexLibrary/app"
	"github.com/pkg/errors"
)

type adminPage struct {
	templateHandler
	data struct {
		User      *app.User
		Tab       string
		Overview  *app.Overview
		WebConfig Config
	}
}

func (a *adminPage) loadShared(s *app.Session) error {
	if s == nil {
		return app.Unauthorized("You do not have access to this page")
	}

	u, err := s.User()
	if err != nil {
		return err
	}

	if !u.Admin {
		return app.Unauthorized("You do not have access to this page")
	}

	a.data.User = u
	a.data.WebConfig = currentConfig
	return nil
}

func (a *adminPage) overview(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {

	a.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		if errHandled(a.loadShared(c.session), w, r) {
			return
		}
		a.data.Tab = "overview"
		overview, err := a.data.User.AsAdmin().Overview()
		if errHandled(err, w, r) {
			return
		}

		a.data.Overview = overview
		err = w.(*templateWriter).execute(a.data)

		if err != nil {
			app.LogError(errors.Wrap(err, "Executing admin template: %s"))
		}
	}
	a.ServeHTTP(w, r, parms)
}

func (a *adminPage) settings(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {

	a.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		if errHandled(a.loadShared(c.session), w, r) {
			return
		}
		a.data.Tab = "settings"
		err := w.(*templateWriter).execute(a.data)

		if err != nil {
			app.LogError(errors.Wrap(err, "Executing admin template: %s"))
		}
	}
	a.ServeHTTP(w, r, parms)
}

func (a *adminPage) logs(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {

	a.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		if errHandled(a.loadShared(c.session), w, r) {
			return
		}
		a.data.Tab = "logs"
		err := w.(*templateWriter).execute(a.data)

		if err != nil {
			app.LogError(errors.Wrap(err, "Executing admin template: %s"))
		}
	}
	a.ServeHTTP(w, r, parms)
}

func (a *adminPage) registration(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {

	a.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		if errHandled(a.loadShared(c.session), w, r) {
			return
		}
		a.data.Tab = "registration"
		err := w.(*templateWriter).execute(a.data)

		if err != nil {
			app.LogError(errors.Wrap(err, "Executing admin template: %s"))
		}
	}
	a.ServeHTTP(w, r, parms)
}
