// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"database/sql/driver"

	"github.com/blevesearch/snowballstem"
	"github.com/blevesearch/snowballstem/arabic"
	"github.com/blevesearch/snowballstem/danish"
	"github.com/blevesearch/snowballstem/dutch"
	"github.com/blevesearch/snowballstem/english"
	"github.com/blevesearch/snowballstem/finnish"
	"github.com/blevesearch/snowballstem/french"
	"github.com/blevesearch/snowballstem/german"
	"github.com/blevesearch/snowballstem/hungarian"
	"github.com/blevesearch/snowballstem/italian"
	"github.com/blevesearch/snowballstem/norwegian"
	"github.com/blevesearch/snowballstem/portuguese"
	"github.com/blevesearch/snowballstem/romanian"
	"github.com/blevesearch/snowballstem/russian"
	"github.com/blevesearch/snowballstem/spanish"
	"github.com/blevesearch/snowballstem/swedish"
	"github.com/blevesearch/snowballstem/tamil"
	"github.com/blevesearch/snowballstem/turkish"
	"golang.org/x/text/language"
)

//https://blog.golang.org/matchlang

var languages = []language.Tag{
	language.English,
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
}

var languageMatcher = language.NewMatcher(languages)

type Language struct {
	language.Tag
}

// newLanguage creates a new language type for language specific text processing, user input should probably
// use MatchLanguage
func newLanguage(lan string) (Language, error) {
	tag, err := language.Parse(lan)
	if err != nil {
		return Language{}, err
	}
	return Language{tag}, nil
}

// MatchLanguage returns a language matched against only the supported list of languages, or returns
// the instance default language if none match
func MatchLanguage(lans ...string) Language {
	tag, _ := language.MatchStrings(languageMatcher, lans...)
	return Language{tag}
}

// Scan implements the Scanner interface.
func (l *Language) Scan(value interface{}) error {
	//FIXME
	val, err := driver.String.ConvertValue(value)
	if err != nil {
		return err
	}

	tag, err := language.Parse(val.(string))
	if err != nil {
		return err
	}

	l.Tag = tag
	return nil
}

// Value implements the driver Valuer interface.
func (l Language) Value() (driver.Value, error) {
	return l.String(), nil
}

// Stem returns the stem of the given word for the given language
func (l Language) Stem(word string) string {
	env := snowballstem.NewEnv(word)
	switch l.Tag {
	case language.Arabic:
		arabic.Stem(env)
	case language.Danish:
		danish.Stem(env)
	case language.Dutch:
		dutch.Stem(env)
	case language.English:
		english.Stem(env)
	case language.Finnish:
		finnish.Stem(env)
	case language.French:
		french.Stem(env)
	case language.German:
		german.Stem(env)
	case language.Hungarian:
		hungarian.Stem(env)
	case language.Italian:
		italian.Stem(env)
	case language.Norwegian:
		norwegian.Stem(env)
	case language.Portuguese:
		portuguese.Stem(env)
	case language.Romanian:
		romanian.Stem(env)
	case language.Russian:
		russian.Stem(env)
	case language.Spanish:
		spanish.Stem(env)
	case language.Swedish:
		swedish.Stem(env)
	case language.Tamil:
		tamil.Stem(env)
	case language.Turkish:
		turkish.Stem(env)
	default:
		// no language found return unstemmed word
		return word
	}

	return env.Current()
}
