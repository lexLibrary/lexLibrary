// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
)

type documentInput struct {
	Title    *string  `json:"title,omitempty"`
	Content  *string  `json:"content,omitempty"`
	Language *string  `json:"language,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Version  *int     `json:"version,omitempty"`
}

func documentNew(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	input := &documentInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r) {
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r) {
		return
	}

	if input.Title == nil {
		errHandled(app.NewFailure("title is required"), w, r)
		return
	}

	lan := c.language

	if input.Language != nil {
		lan = languageFromString(*input.Language)
	}

	doc, err := u.NewDocument(*input.Title, lan)
	if errHandled(err, w, r) {
		return
	}

	respond(w, created(doc))
}

func draftSave(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	input := &documentInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r) {
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r) {
		return
	}

	var title, content string

	if input.Title != nil {
		title = *input.Title
	}

	if input.Content != nil {
		content = *input.Content
	}

	if input.Version == nil {
		errHandled(app.NewFailure("Version is required"), w, r)
		return
	}

	lan := c.language
	if input.Language != nil {
		lan = languageFromString(*input.Language)
	}

	id, err := data.IDFromString(c.params.ByName("id"))
	if err != nil {
		notFound(w, r)
	}

	draft, err := u.Draft(id, lan)
	if errHandled(err, w, r) {
		return
	}

	if errHandled(draft.Save(title, content, input.Tags, *input.Version), w, r) {
		return
	}

	respond(w, success(nil))
}

// func draftNew(w http.ResponseWriter, r *http.Request, c ctx) {
// 	if c.session == nil {
// 		unauthorized(w, r)
// 		return
// 	}

// 	u, err := c.session.User()
// 	if errHandled(err, w, r) {
// 		return
// 	}

// 	id, err := data.IDFromString(c.params.ByName("id"))
// 	if err != nil {
// 		notFound(w, r)
// 	}
// }
