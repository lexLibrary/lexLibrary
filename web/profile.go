// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/lexLibrary/lexLibrary/app"
)

type profilePage struct {
	*templateHandler
	self bool
}

type profileData struct {
	User  *app.PublicProfile
	Stats app.UserStats
	Tab   string
	Self  bool
	Base  string
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
		serveImage(w, r, u.ProfileImageDraft(), false)
		return
	}

	serveImage(w, r, u.ProfileImage(), false)
}

func (p profilePage) data(c ctx) (*profileData, error) {

	var u *app.PublicProfile
	var private *app.User
	var err error
	base := "/profile"
	self := p.self

	if p.self {
		if c.session == nil {
			return nil, app.Unauthorized("You do not have access to this user")
		}
		private, err = c.session.User()
		if err != nil {
			return nil, err
		}

		u = &private.PublicProfile
	} else {
		u, err = app.UserGet(c.params.ByName("username"))
		if err != nil {
			return nil, err
		}
		base = "/user/" + u.Username

		if c.session != nil && c.session.UserID == u.ID {
			self = true
		}
	}

	stats, err := u.Stats()
	if err != nil {
		return nil, err
	}

	return &profileData{
		User:  u,
		Stats: stats,
		Self:  self,
		Base:  base,
	}, nil
}

func (p *profilePage) documents() httprouter.Handle {
	return p.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := p.data(c)
		if errHandled(err, w, r) {
			return
		}

		tData.Tab = "documents"
		w.(*templateWriter).execute(tData)
	})
}

func (p *profilePage) readLater() httprouter.Handle {
	return p.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := p.data(c)
		if !tData.Self {
			notFound(w, r)
			return
		}
		if errHandled(err, w, r) {
			return
		}

		tData.Tab = "readLater"
		w.(*templateWriter).execute(tData)
	})
}

func (p *profilePage) comments() httprouter.Handle {
	return p.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := p.data(c)
		if errHandled(err, w, r) {
			return
		}

		tData.Tab = "comments"
		w.(*templateWriter).execute(tData)
	})
}

func (p *profilePage) history() httprouter.Handle {
	return p.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := p.data(c)
		if !tData.Self {
			notFound(w, r)
			return
		}
		if errHandled(err, w, r) {
			return
		}

		tData.Tab = "history"
		w.(*templateWriter).execute(tData)
	})
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

func (p *profileEditPage) root() httprouter.Handle {
	return p.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {
		err := p.loadShared(c.session)
		if errHandled(err, w, r) {
			return
		}
		p.data.Tab = "profile"

		w.(*templateWriter).execute(p.data)
	})
}

func (p *profileEditPage) account() httprouter.Handle {
	return p.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {
		err := p.loadShared(c.session)
		if errHandled(err, w, r) {
			return
		}
		p.data.Tab = "account"

		w.(*templateWriter).execute(p.data)
	})
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

func profileRemoveImage(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r) {
		return
	}

	if errHandled(u.RemoveProfileImage(), w, r) {
		return
	}

	respond(w, success(nil))
}
