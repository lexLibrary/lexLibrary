// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"
	"path"

	"github.com/julienschmidt/httprouter"
	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
)

type editorPage struct {
	*templateHandler
}

type editorData struct {
	User  *app.User
	Draft *app.DocumentDraft
	Page  string

	Languages []app.Language
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
		User:      u,
		Languages: app.LanguagesSupported,
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

func (e *editorPage) edit() httprouter.Handle {
	return e.handle(func(w http.ResponseWriter, r *http.Request, c ctx) {

		tData, err := e.data(c.session)
		if errHandled(err, w, r) {
			return
		}

		tData.Page = "edit"
		id, err := data.IDFromString(c.params.ByName("id"))
		if err != nil {
			notFound(w, r)
		}

		draft, err := tData.User.Draft(id)
		if errHandled(err, w, r) {
			return
		}

		ogSlug := c.params.ByName("slug")
		slug := draft.Slug()

		if ogSlug != slug {
			if ogSlug == "" {
				http.Redirect(w, r, path.Join(r.URL.Path, slug), http.StatusPermanentRedirect)
			} else {
				// if title mis-matches, use a temporary redirect in case they change the title back
				// to something they had previously, so we prevent a loop of redirects

				http.Redirect(w, r, path.Join(path.Dir(r.URL.Path), slug), http.StatusTemporaryRedirect)
			}
			return
		}

		tData.Draft = draft

		w.(*templateWriter).execute(tData)
	})
}
