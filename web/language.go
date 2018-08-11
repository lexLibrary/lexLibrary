// Copyright (c) 2017-2018 Townsourced Inc.

package web

import (
	"net/http"

	"golang.org/x/text/language"
)

var lanMatcher = language.NewMatcher([]language.Tag{
	language.English, // The first language is used as fallback. TODO: set to server setting default
	language.Arabic,
	language.Danish,
	language.Dutch,
	language.English,
	language.Finnish,
	language.French,
	language.German,
	language.Hungarian,
	language.Italian,
	language.Norwegian,
	language.Portuguese,
	language.Romanian,
	language.Russian,
	language.Spanish,
	language.Swedish,
	language.Tamil,
	language.Turkish,
})

func languageFromRequest(r *http.Request) language.Tag {
	lanCookie := ""
	lang, err := r.Cookie("lang")
	if err == nil {
		lanCookie = lang.String()
	}
	accept := r.Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(lanMatcher, lanCookie, accept)

	return tag
}

func languageFromString(lan string) language.Tag {
	tag, _ := language.MatchStrings(lanMatcher, lan)
	return tag
}
