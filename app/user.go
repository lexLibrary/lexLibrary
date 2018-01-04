// Copyright (c) 2017 Townsourced Inc.

package app

import (
	"database/sql"
	"strings"
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
	//AuthType...
)

// user constants
const (
	UserMaxNameLength = 64 // This is pretty arbitrary but there should probably be some limit
)

var (
	// ErrLogonFailure is when a user fails a login attempt
	ErrLogonFailure = NewFailure("Invalid user and / or password")
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
	sqlUserIDFromUsername = data.NewQuery(`select id from users where username = {{arg "username"}}`)
	sqlUserFromID         = data.NewQuery(`
		select id, username, first_name, last_name, auth_type, active, version,	updated, created 
		from users where id = {{arg "id"}}`)
)

// UserNew creates a new user
func UserNew(username, firstName, lastName, password string) (*User, error) {
	u := &User{
		ID:        xid.New(),
		Username:  strings.ToLower(username),
		FirstName: firstName,
		LastName:  lastName,
		AuthType:  AuthTypePassword,
		Active:    true,
		Version:   0,
		Created:   time.Now(),
		Updated:   time.Now(),
	}

	err := u.validate()
	if err != nil {
		return nil, err
	}

	err = validatePassword(password)
	if err != nil {
		return nil, err
	}

	passVer := len(passwordVersions) - 1

	hash, err := passwordVersions[passVer].hash(password)
	if err != nil {
		return nil, err
	}

	u.PasswordVersion = passVer
	u.Password = hash

	exists, err := u.exists()
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, NewFailure("A user with the username %s already exists", u.Username)
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

func (u *User) exists() (bool, error) {
	var id xid.ID
	err := sqlUserIDFromUsername.QueryRow(sql.Named("username", u.Username)).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (u *User) validate() error {
	if u.Username == "" {
		return NewFailure("A username is required")
	}

	if !urlify(u.Username).is() {
		return NewFailure("A username can only contain letters, numbers and dashes")
	}

	if len(u.FirstName) > UserMaxNameLength {
		return NewFailure("First name must be less than %d characters", UserMaxNameLength)
	}
	if len(u.LastName) > UserMaxNameLength {
		return NewFailure("Last name must be less than %d characters", UserMaxNameLength)
	}

	if u.AuthType != AuthTypePassword {
		return NewFailure("Invalid user Authentication Type")
	}

	return nil
}
