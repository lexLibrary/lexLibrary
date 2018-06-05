// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/lexLibrary/lexLibrary/data"
	"github.com/pkg/errors"
)

const (
	sessionValDelim = "@"
	cookieName      = "lexlibrary"
)

type sessionInput struct {
	userInput
	RememberMe bool `json:"rememberMe,omitempty"`
}

// rate limit login attempts
var logonDelay = app.RateDelay{
	Type:   "login",
	Limit:  10,
	Delay:  5 * time.Second,
	Period: 5 * time.Minute,
	Max:    1 * time.Minute,
}

func loginTemplate(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session != nil {
		// already logged in, redirect to home
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	w.(*templateWriter).execute(map[string]bool{
		"AllowSignup": app.SettingMust("AllowPublicSignups").Bool(),
	})
}

func signupTemplate(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session != nil {
		// already logged in, redirect to home
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if !app.SettingMust("AllowPublicSignups").Bool() {
		notFound(w, r)
		return
	}
	w.(*templateWriter).execute(nil)
}

// get a session from the request
func session(r *http.Request) (*app.Session, error) {
	// must iter through all cookies because you can have
	// multiple cookies with the same name
	// the cookie is valid only if the name matches AND it has a value
	cookies := r.Cookies()
	cValue := ""

	for i := range cookies {
		if cookies[i].Name == cookieName {
			if cookies[i].Value != "" {
				cValue = cookies[i].Value
			}
		}
	}

	if cValue == "" {
		return nil, nil
	}

	ids := strings.Split(cValue, sessionValDelim)

	userID, err := data.IDFromString(ids[0])
	if err != nil {
		return nil, nil
	}

	s, err := app.SessionGet(userID, ids[1])
	if err == app.ErrSessionInvalid {
		return nil, nil
	}
	return s, err
}

func handleCSRF(w http.ResponseWriter, r *http.Request, s *app.Session) error {
	if s == nil {
		return nil
	}

	if r.Method != "GET" {
		reqToken := r.Header.Get("X-CSRFToken")
		if reqToken != s.CSRFToken && s.Valid {
			return app.NewFailure("Invalid CSRFToken.  Your session may be invalid.  Try logging in again.")
		}

		return nil
	}

	// FIXME:
	err := s.CycleCSRF()
	if err != nil {
		return err
	}

	if w, ok := w.(*templateWriter); ok {
		w.CSRFToken = s.CSRFToken
	}

	//Get requests, put CSRF token in header
	w.Header().Add("X-CSRFToken", s.CSRFToken)

	return nil
}

func setSession(w http.ResponseWriter, r *http.Request, u *app.User, rememberMe bool) (*app.Session, error) {
	expires := time.Time{}

	if rememberMe {
		expires = time.Now().AddDate(0, 0, app.SettingMust("RememberSessionDays").Int())
	}

	s, err := u.NewSession(expires, ipAddress(r), r.UserAgent())
	if err != nil {
		return nil, err
	}

	key := s.UserID.String() + sessionValDelim + s.ID
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    key,
		HttpOnly: true,
		Path:     "/",
		Secure:   isSSL,
		Expires:  expires,
	}

	http.SetCookie(w, cookie)
	w.Header().Add("X-CSRFToken", s.CSRFToken)
	return s, nil
}

func expireSessionCookie(w http.ResponseWriter, r *http.Request, s *app.Session) {
	cookie, err := r.Cookie(cookieName)

	if err != http.ErrNoCookie {
		key := s.UserID.String() + sessionValDelim + s.ID
		if cookie.Value == key {
			cookie := &http.Cookie{
				Name:     cookieName,
				Value:    "",
				HttpOnly: true,
				Path:     "/",
				Secure:   isSSL,
				MaxAge:   0,
			}

			http.SetCookie(w, cookie)
		}
	}
}

func sessionLogin(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session != nil {
		//If previous session still exists, log out so it can't be used again
		go func(session *app.Session) {
			err := session.Logout()
			if err != nil {
				app.LogError(errors.Wrap(err, "Logging out session when trying to log into a new session"))
			}
		}(c.session)
	}

	input := &sessionInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r) {
		return
	}

	if input.Username == nil {
		errHandled(app.NewFailure("A username is required"), w, r)
		return
	}

	if input.Password == nil {
		errHandled(app.NewFailure("You must specify a password"), w, r)
		return
	}

	// rate limit login requests
	if errHandled(logonDelay.Attempt(ipAddress(r)), w, r) {
		return
	}

	u, err := app.Login(*input.Username, *input.Password)
	if err == app.ErrPasswordExpired {
		respond(w, success(map[string]bool{
			"expired": true,
		}))
		return
	}
	if errHandled(err, w, r) {
		return
	}

	_, err = setSession(w, r, u, input.RememberMe)
	if errHandled(err, w, r) {
		return
	}

	respond(w, created(u))
}

func sessionLogout(w http.ResponseWriter, r *http.Request, c ctx) {
	if c.session == nil {
		respond(w, success(nil))
		return
	}

	expireSessionCookie(w, r, c.session)

	if errHandled(c.session.Logout(), w, r) {
		return
	}

	respond(w, success(nil))
}
