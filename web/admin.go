// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
	"github.com/pkg/errors"
)

type adminPage struct {
	templateHandler
}

type adminData struct {
	User      *app.User
	Tab       string
	Overview  *app.Overview
	WebConfig Config
	Log       struct {
		Logs  []*app.Log
		Pager pager
		Entry *app.Log
	}
	Settings []app.Setting
}

func (a *adminPage) data(s *app.Session) (*adminData, error) {
	if s == nil {
		return nil, app.Unauthorized("You do not have access to this page")
	}

	u, err := s.User()
	if err != nil {
		return nil, err
	}

	if !u.Admin {
		return nil, app.Unauthorized("You do not have access to this page")
	}

	return &adminData{
		User: u,
	}, nil

}

func (a *adminPage) overview(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {
	a.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := a.data(c.session)
		if errHandled(err, w, r) {
			return
		}

		tData.Tab = "overview"
		tData.WebConfig = currentConfig
		overview, err := tData.User.AsAdmin().Overview()
		if errHandled(err, w, r) {
			return
		}

		tData.Overview = overview
		err = w.(*templateWriter).execute(tData)

		if err != nil {
			app.LogError(errors.Wrap(err, "Executing admin template: %s"))
		}
	}
	a.ServeHTTP(w, r, parms)
}

func (a *adminPage) settings(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {

	a.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := a.data(c.session)
		if errHandled(err, w, r) {
			return
		}

		tData.Tab = "settings"
		settings, err := app.Settings(tData.User)
		if errHandled(err, w, r) {
			return
		}
		tData.Settings = settings
		err = w.(*templateWriter).execute(tData)

		if err != nil {
			app.LogError(errors.Wrap(err, "Executing admin template: %s"))
		}
	}
	a.ServeHTTP(w, r, parms)
}

func (a *adminPage) logs(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {

	a.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := a.data(c.session)
		if errHandled(err, w, r) {
			return
		}

		var id data.ID
		var search string
		tData.Tab = "logs"

		if c.params.ByName("id") != "" {
			id, err = data.IDFromString(c.params.ByName("id"))
			if err != nil {
				notFound(w, r)
				return
			}
		} else {
			search = r.URL.Query().Get("search")
			id, _ = data.IDFromString(search)
		}

		if !id.IsNil() {
			log, err := app.LogGetByID(id)
			if errHandled(err, w, r) {
				return
			}
			tData.Log.Entry = log
		} else {
			var logs []*app.Log
			total := 0
			pgr := newPager(r.URL, 30)

			if search != "" {
				logs, total, err = app.LogSearch(search, pgr.Offset(), pgr.PageSize())
			} else {
				logs, total, err = app.LogGet(pgr.Offset(), pgr.PageSize())
			}
			if errHandled(err, w, r) {
				return
			}

			pgr.SetTotal(total)
			tData.Log.Pager = pgr
			tData.Log.Logs = logs

		}
		err = w.(*templateWriter).execute(tData)

		if err != nil {
			app.LogError(errors.Wrap(err, "Executing admin template: %s"))
		}
	}
	a.ServeHTTP(w, r, parms)
}

func (a *adminPage) registration(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {

	a.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := a.data(c.session)
		if errHandled(err, w, r) {
			return
		}

		tData.Tab = "registration"
		err = w.(*templateWriter).execute(tData)

		if err != nil {
			app.LogError(errors.Wrap(err, "Executing admin template: %s"))
		}
	}
	a.ServeHTTP(w, r, parms)
}
