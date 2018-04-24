// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"database/sql"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
	"github.com/pkg/errors"
)

// Session is a user login session to Lex Library
type Session struct {
	ID        string
	UserID    data.ID
	Valid     bool
	Expires   time.Time
	IPAddress string
	UserAgent string
	CSRFToken string
	CSRFDate  time.Time
	Created   time.Time
	Updated   time.Time

	user *User // cached user record
}

const (
	sessionMaxDaysRemembered = 365
	csrfTokenMaxAge          = 15 * time.Minute // max age of a csrf token before it's reset
)

var (
	// ErrSessionInvalid is returned when a sesssion is invalid or expired
	ErrSessionInvalid = NewFailure("Invalid or expired session")
	// ErrLogonFailure is when a user fails a login attempt
	ErrLogonFailure = NewFailure("Invalid user and / or password")
	// ErrPasswordExpired is when a user's password has expired
	ErrPasswordExpired = NewFailure("Your password has expired.  Please set a new one.")
)

var (
	sqlSessionInsert = data.NewQuery(`insert into sessions (
		id,
		user_id,
		valid,
		expires,
		ip_address,
		user_agent,
		csrf_token,
		csrf_date,
		created,
		updated
	) values (
		{{arg "id"}},
		{{arg "user_id"}},
		{{arg "valid"}},
		{{arg "expires"}},
		{{arg "ip_address"}},
		{{arg "user_agent"}},
		{{arg "csrf_token"}},
		{{arg "csrf_date"}},
		{{arg "created"}},
		{{arg "updated"}}
	)`)
	sqlSessionSetValid = data.NewQuery(`update sessions set valid = {{arg "valid"}} where id = {{arg "id"}}`)
	sqlSessionSetCSRF  = data.NewQuery(`update sessions 
		set csrf_token = {{arg "csrf_token"}}, csrf_date = {{arg "csrf_date"}} where id = {{arg "id"}}`)
	sqlSessionGet = data.NewQuery(`select id, user_id, valid, expires, csrf_token, csrf_date 
		from sessions where id = {{arg "id"}} and user_id = {{arg "user_id"}}`)
	sqlSessionInvalidateAll = data.NewQuery(`
		update sessions set valid = {{FALSE}} 
		where user_id = {{arg "user_id"}} 
		and valid <> {{FALSE}}
		and expires >= {{NOW}}
	`)
)

// Login logs a new user into Lex Library.
func Login(username string, password string) (*User, error) {
	if username == "" || len(password) < passwordMinLength {
		return nil, ErrLogonFailure
	}

	u, err := userFromUsername(nil, username)

	if err == ErrUserNotFound {
		return nil, ErrLogonFailure
	}
	if err != nil {
		return nil, err
	}

	if !u.Active {
		return nil, ErrLogonFailure
	}

	if u.AuthType == AuthTypePassword {
		err = passwordVersions[u.passwordVersion].compare(password, u.password)
		if err != nil {
			if err != ErrPasswordMismatch {
				LogError(err)
			}
			return nil, ErrLogonFailure
		}
		if u.PasswordExpiration.Valid && u.PasswordExpiration.Time.Before(time.Now()) {
			return nil, ErrPasswordExpired
		}
	} else {
		return nil, errors.Errorf("The user %s is stored in the database with an invalid authentication type."+
			" This could mean that an older version of Lex Library is running on a newer version of the database.",
			u.Username)
	}
	return u, nil
}

// SessionNew generates a new session for the passed in user
func SessionNew(user *User, expires time.Time, ipAddress, userAgent string) (*Session, error) {
	if expires.IsZero() {
		expires = time.Now().AddDate(0, 0, 3)
	}

	s := &Session{
		ID:        Random(128),
		UserID:    user.ID,
		CSRFToken: Random(256),
		CSRFDate:  time.Now(),
		Valid:     true,
		Expires:   expires,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Created:   time.Now(),
		Updated:   time.Now(),
	}

	err := s.insert()

	if err != nil {
		return nil, err
	}
	s.user = user

	return s, nil
}

// SessionGet retrieves a session
func SessionGet(userID data.ID, sessionID string) (*Session, error) {
	s := &Session{}
	err := sqlSessionGet.QueryRow(sql.Named("id", sessionID), sql.Named("user_id", userID)).
		Scan(
			&s.ID,
			&s.UserID,
			&s.Valid,
			&s.Expires,
			&s.CSRFToken,
			&s.CSRFDate,
		)

	if err == sql.ErrNoRows {
		return nil, ErrSessionInvalid
	}

	if err != nil {
		return nil, err
	}

	if !s.Valid || s.Expires.Before(time.Now()) {
		return nil, ErrSessionInvalid
	}

	return s, nil
}

func (s *Session) insert() error {
	_, err := sqlSessionInsert.Exec(
		sql.Named("id", s.ID),
		sql.Named("user_id", s.UserID),
		sql.Named("valid", s.Valid),
		sql.Named("expires", s.Expires),
		sql.Named("ip_address", s.IPAddress),
		sql.Named("user_agent", s.UserID),
		sql.Named("csrf_token", s.CSRFToken),
		sql.Named("csrf_date", s.CSRFDate),
		sql.Named("created", s.Created),
		sql.Named("updated", s.Updated),
	)
	return err
}

// Logout logs a session out
func (s *Session) Logout() error {
	s.Valid = false
	_, err := sqlSessionSetValid.Exec(sql.Named("valid", s.Valid), sql.Named("id", s.ID))
	return err
}

// User returns the user for the given login session
func (s *Session) User() (*User, error) {
	if s.user != nil {
		return s.user, nil
	}
	u, err := userFromID(nil, s.UserID)
	if err == ErrUserNotFound {
		return nil, ErrSessionInvalid
	}
	if err != nil {
		return nil, err
	}

	if !u.Active {
		return nil, ErrSessionInvalid
	}

	s.user = u

	return u, nil
}

// CycleCSRF will generate a new CRSF token if it is too old, and update the session with it
// allows CSRF token to change more than once per session if need be
func (s *Session) CycleCSRF() error {
	if s.CSRFDate.Add(csrfTokenMaxAge).After(time.Now()) {
		return nil
	}

	s.CSRFToken = Random(256)
	s.CSRFDate = time.Now()
	_, err := sqlSessionSetCSRF.Exec(sql.Named("csrf_token", s.CSRFToken), sql.Named("csrf_date", s.CSRFDate),
		sql.Named("id", s.ID))
	return err
}
