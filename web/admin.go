// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
)

type adminPage struct {
	*templateHandler
}

type adminData struct {
	User      *app.User
	Admin     *app.Admin
	Tab       string
	Overview  *app.Overview
	WebConfig Config
	Log       struct {
		Logs  []*app.Log
		Pager pager
		Entry *app.Log
	}
	Settings     []app.Setting
	Registration struct {
		New    bool
		Tokens []*app.RegistrationToken
		Pager  pager
		All    bool
		Single *app.RegistrationToken
	}
	Users []*app.PublicProfile
}

func (a *adminPage) data(s *app.Session) (*adminData, error) {
	if s == nil {
		return nil, app.Unauthorized("You do not have access to this page")
	}

	u, err := s.User()
	if err != nil {
		return nil, err
	}
	admin, err := u.Admin()
	if err != nil {
		return nil, err
	}

	return &adminData{
		User:  u,
		Admin: admin,
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
		overview, err := tData.Admin.Overview()
		if errHandled(err, w, r) {
			return
		}

		tData.Overview = overview
		w.(*templateWriter).execute(tData)
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
		settings, err := tData.Admin.Settings()
		if errHandled(err, w, r) {
			return
		}
		tData.Settings = settings
		w.(*templateWriter).execute(tData)
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
		w.(*templateWriter).execute(tData)
	}
	a.ServeHTTP(w, r, parms)
}

func (a *adminPage) registration(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {

	a.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := a.data(c.session)
		if errHandled(err, w, r) {
			return
		}

		var tokens []*app.RegistrationToken
		tData.Tab = "registration"
		pgr := newPager(r.URL, 20)
		_, ok := r.URL.Query()["all"]
		tData.Registration.All = ok

		tokens, total, err := tData.Admin.RegistrationTokenList(!tData.Registration.All,
			pgr.Offset(), pgr.PageSize())
		if errHandled(err, w, r) {
			return
		}
		pgr.SetTotal(total)
		tData.Registration.Pager = pgr
		tData.Registration.Tokens = tokens

		w.(*templateWriter).execute(tData)
	}
	a.ServeHTTP(w, r, parms)
}

func (a *adminPage) registrationNew(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {

	a.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := a.data(c.session)
		if errHandled(err, w, r) {
			return
		}

		tData.Tab = "registration"
		tData.Registration.New = true

		w.(*templateWriter).execute(tData)
	}
	a.ServeHTTP(w, r, parms)
}

func (a *adminPage) registrationGet(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {

	a.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := a.data(c.session)
		if errHandled(err, w, r) {
			return
		}

		tData.Tab = "registration"
		token, err := tData.Admin.RegistrationToken(c.params.ByName("token"))
		if errHandled(err, w, r) {
			return
		}
		tData.Registration.Single = token

		w.(*templateWriter).execute(tData)
	}
	a.ServeHTTP(w, r, parms)
}

func (a *adminPage) users(w http.ResponseWriter, r *http.Request, parms httprouter.Params) {

	a.handler = func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := a.data(c.session)
		if errHandled(err, w, r) {
			return
		}

		tData.Tab = "users"
		// qry := r.URL.Query()

		w.(*templateWriter).execute(tData)
	}
	a.ServeHTTP(w, r, parms)
}
