// Copyright (c) 2017-2018 Townsourced Inc.

package web

import (
	"net/http"
	"strings"

	"github.com/lexLibrary/lexLibrary/app"
)

type groupInput struct {
	Name *string
}

func groupGet(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r) {
		return
	}
	query := r.URL.Query()
	search := query.Get("search")

	if strings.TrimSpace(search) != "" {
		groups, err := u.GroupSearch(search)
		if errHandled(err, w, r) {
			return
		}
		respond(w, success(groups))
		return
	}

	respond(w, success(nil))
}

func groupCreate(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	input := &groupInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r) {
		return
	}

	if input.Name == nil {
		errHandled(app.NewFailure("Name is required"), w, r)
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r) {
		return
	}

	g, err := u.NewGroup(*input.Name)
	if errHandled(err, w, r) {
		return
	}

	respond(w, created(g))
}
