// Copyright (c) 2018 Townsourced Inc.

package app

import (
	"database/sql"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
	"github.com/rs/xid"
)

// Session is a user login session to Lex Library
type Session struct {
	ID        string
	UserID    xid.ID
	Valid     bool
	Expires   time.Time
	IPAddress string
	UserAgent string
	CSRFToken string
	Created   time.Time
	Updated   time.Time
}

// ErrSessionInvalid is returned when a sesssion is invalid or expired
var ErrSessionInvalid = NewFailure("Invalid or expired session")

var (
	sqlSessionInsert = data.NewQuery(`insert into sessions (
		id,
		user_id,
		valid,
		expires,
		ip_address,
		user_agent,
		csrf_token,
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
		{{arg "created"}},
		{{arg "updated"}}
	)`)
	sqlSessionSetValid = data.NewQuery(`update sessions set valid = {{arg "valid"}} where id = {{arg "id"}}`)
	sqlSessionSetCSRF  = data.NewQuery(`update sessions set csrf_token = {{arg "csrf_token"}} where id = {{arg "id"}}`)
)

// SessionNew generates a new session for the passed in user
func SessionNew(user *User, expires time.Time, ipAddress, userAgent string) (*Session, error) {
	if expires.IsZero() {
		expires = time.Now().AddDate(0, 0, 3)
	}

	s := &Session{
		ID:        Random(128),
		UserID:    user.ID,
		CSRFToken: Random(256),
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
	u := &User{}
	err := sqlUserFromID.QueryRow(sql.Named("id", s.UserID)).Scan(
		&u.ID,
		&u.Username,
		&u.FirstName,
		&u.LastName,
		&u.AuthType,
		&u.Active,
		&u.Version,
		&u.Updated,
		&u.Created)
	if err == sql.ErrNoRows {
		return nil, ErrSessionInvalid
	}
	if err != nil {
		return nil, err
	}

	if !u.Active {
		return nil, ErrSessionInvalid
	}

	return u, nil
}

// ResetCSRF will generate a new CRSF token, an update the session with it
// allows CSRF token to change more than once per session if need be
func (s *Session) ResetCSRF() error {
	s.CSRFToken = Random(256)
	_, err := sqlSessionSetCSRF.Exec(sql.Named("csrf_token", s.CSRFToken), sql.Named("id", s.ID))
	return err
}
