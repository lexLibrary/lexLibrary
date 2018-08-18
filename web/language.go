// Copyright (c) 2017-2018 Townsourced Inc.

package web

import (
	"net/http"

	"github.com/lexLibrary/lexLibrary/app"
)

func languageFromRequest(r *http.Request) app.Language {
	lanCookie := ""
	lang, err := r.Cookie("lang")
	if err == nil {
		lanCookie = lang.String()
	}
	accept := r.Header.Get("Accept-Language")

	return app.MatchLanguage(lanCookie, accept)
}
