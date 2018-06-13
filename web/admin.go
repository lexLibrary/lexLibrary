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
	Users struct {
		Active   bool
		LoggedIn bool
		All      bool
		Search   string
		Users    []*app.InstanceUser
		User     *app.InstanceUser
		Pager    pager
	}
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

func (a *adminPage) overview() httprouter.Handle {
	return a.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {
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
	})
}

func (a *adminPage) settings() httprouter.Handle {
	return a.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {
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
	})
}

func (a *adminPage) logs() httprouter.Handle {
	return a.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {
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
	})
}

func (a *adminPage) registration() httprouter.Handle {
	return a.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := a.data(c.session)
		if errHandled(err, w, r) {
			return
		}

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
	})
}

func (a *adminPage) registrationNew() httprouter.Handle {
	return a.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := a.data(c.session)
		if errHandled(err, w, r) {
			return
		}

		tData.Tab = "registration"
		tData.Registration.New = true

		w.(*templateWriter).execute(tData)
	})
}

func (a *adminPage) registrationGet() httprouter.Handle {
	return a.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {
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
	})
}

func (a *adminPage) users() httprouter.Handle {
	return a.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := a.data(c.session)
		if errHandled(err, w, r) {
			return
		}

		tData.Tab = "users"
		qry := r.URL.Query()
		pgr := newPager(r.URL, 20)

		_, tData.Users.Active = qry["active"]
		_, tData.Users.LoggedIn = qry["loggedin"]
		tData.Users.Search = qry.Get("search")
		tData.Users.All = (!tData.Users.Active && !tData.Users.LoggedIn)

		users, total, err := tData.Admin.InstanceUsers(tData.Users.Active, tData.Users.LoggedIn,
			tData.Users.Search, pgr.Offset(), pgr.PageSize())
		if errHandled(err, w, r) {
			return
		}
		pgr.SetTotal(total)
		tData.Users.Pager = pgr
		tData.Users.Users = users

		w.(*templateWriter).execute(tData)
	})
}

func (a *adminPage) user() httprouter.Handle {
	return a.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := a.data(c.session)
		if errHandled(err, w, r) {
			return
		}

		tData.Tab = "users"
		u, err := tData.Admin.InstanceUser(c.params.ByName("username"))
		if errHandled(err, w, r) {
			return
		}

		tData.Users.User = u

		w.(*templateWriter).execute(tData)
	})
}

type adminUserInput struct {
	Active *bool `json:"active"`
	Admin  *bool `json:"admin"`
}

func adminUserUpdate(w http.ResponseWriter, r *http.Request, c ctx) {
	input := &adminUserInput{}
	if errHandled(parseInput(r, input), w, r) {
		return
	}

	if c.session == nil {
		errHandled(app.Unauthorized("You do not have access to this page"), w, r)
		return
	}

	admin, err := c.session.Admin()
	if errHandled(err, w, r) {
		return
	}

	username := c.params.ByName("username")

	if input.Active != nil {
		if errHandled(admin.SetUserActive(username, *input.Active), w, r) {
			return
		}
	}

	if input.Admin != nil {
		if errHandled(admin.SetUserAdmin(username, *input.Admin), w, r) {
			return
		}
	}

}
