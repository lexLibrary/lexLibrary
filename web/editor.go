// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/lexLibrary/lexLibrary/app"
)

type editorPage struct {
	*templateHandler
}

type editorData struct {
	User *app.User
	Page string
}

func (e *editorPage) data(s *app.Session) (*editorData, error) {
	if s == nil {
		return nil, app.Unauthorized("You do not have access to this page")
	}

	u, err := s.User()
	if err != nil {
		return nil, err
	}
	return &editorData{
		User: u,
	}, nil
}

func (e *editorPage) newDocument() httprouter.Handle {
	return e.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {
		tData, err := e.data(c.session)
		if errHandled(err, w, r) {
			return
		}

		tData.Page = "new"

		w.(*templateWriter).execute(tData)
	})
}
