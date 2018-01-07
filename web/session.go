// Copyright (c) 2018 Townsourced Inc.
package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/pkg/errors"
	"github.com/rs/xid"
)

const (
	sessionValDelim = "_"
	cookieName      = "lexlibrary"
)

func loginTemplate(w http.ResponseWriter, r *http.Request, c ctx) {
	err := w.(*templateWriter).execute(struct {
		Test string
	}{
		Test: "test string",
	})

	if err != nil {
		app.LogError(errors.Wrap(err, "Executing login template: %s"))
	}
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

	userID, err := xid.FromString(ids[0])
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
		//TODO: Consider resetting CSRF token regularily?
		// On each update? Once a day?
	}

	//Get requests, put CSRF token in header
	w.Header().Add("X-CSRFToken", s.CSRFToken)

	return nil
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, u *app.User, rememberMe bool) error {
	expires := time.Time{}

	if rememberMe {
		expires = time.Now().AddDate(0, 0, 15)
	}

	s, err := app.SessionNew(u, expires, ipAddress(r), r.UserAgent())
	if err != nil {
		return err
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
	return nil
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
