// Copyright (c) 2017 Townsourced Inc.

package app

import (
	"database/sql"
	"time"

	"github.com/lexLibrary/lexLibrary/data"
	"github.com/rs/xid"
)

// User is a user login to Lex Library
type User struct {
	ID              xid.ID    `json:"id"`
	Username        string    `json:"username"`
	FirstName       string    `json:"firstName"`
	LastName        string    `json:"lastName"`
	AuthType        string    `json:"authType"`
	Password        []byte    `json:"-"`
	PasswordVersion int       `json:"-"`
	Active          bool      `json:"active"`  // whether or not the user is active and can log in
	Version         int       `json:"version"` // version of this record starting with 0
	Updated         time.Time `json:"updated,omitempty"`
	Created         time.Time `json:"created,omitempty"`
}

// AuthType determines the authentication method for a given user
const (
	AuthTypePassword = "password"
)

var (
	sqlUserInsert = data.NewQuery(`insert into users (
		id,
		username, 
		first_name, 
		last_name, 
		auth_type,
		password,
		password_version,
		active,
		version,
		updated, 
		created
	) values (
		{{arg "id"}}, 
		{{arg "username"}}, 
		{{arg "first_name"}}, 
		{{arg "last_name"}}, 
		{{arg "auth_type"}},
		{{arg "password"}},
		{{arg "password_version"}},
		{{arg "active"}},
		{{arg "version"}},
		{{arg "updated"}}, 
		{{arg "created"}}
	)`)
)

// UserNew creates a new user
func UserNew(username, firstName, lastName, authType, password string) (*User, error) {
	// validate username length and authtype
	err := validatePassword(password)
	if err != nil {
		return nil, err
	}

	passVer := len(passwordVersions) - 1

	hash, err := passwordVersions[passVer].hash(password)
	if err != nil {
		return nil, err
	}

	u := &User{
		ID:              xid.New(),
		Username:        username,
		FirstName:       firstName,
		LastName:        lastName,
		AuthType:        authType,
		Password:        hash,
		PasswordVersion: passVer,
		Active:          true,
		Version:         0,
		Updated:         time.Now(),
		Created:         time.Now(),
	}

	err = u.insert()
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (u *User) insert() error {
	_, err := sqlUserInsert.Exec(
		sql.Named("id", u.ID),
		sql.Named("username", u.Username),
		sql.Named("first_name", u.FirstName),
		sql.Named("last_name", u.LastName),
		sql.Named("auth_type", u.AuthType),
		sql.Named("password", u.Password),
		sql.Named("password_version", u.PasswordVersion),
		sql.Named("active", u.Active),
		sql.Named("version", u.Version),
		sql.Named("updated", u.Updated),
		sql.Named("created", u.Created),
	)

	return err
}
